package deployment

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Project interface {
	RuntimeConfigFile() (string, error)
	FindMatchingFrameworkVersion(string, string, *bool) (string, error)
	GetVersionFromDepsJSON() (string, error)
	VersionFromProjFile(string, string, string) (string, error)
	MainPath() (string, error)
}

type FrameworkDependentDeployment struct {
	Project          Project
	Installer        Installer
	FrameworkName    string
	FrameworkVersion string
	ApplyPatches     *bool
	DepDir           string
}

type SelfContainedDeployment struct{}

type SourceDeployment struct {
	Project   Project
	Installer Installer
	DepDir    string
}

type Installer interface {
	FetchDependency(libbuildpack.Dependency, string) error
	InstallDependency(libbuildpack.Dependency, string) error
	InstallOnlyVersion(string, string) error
}

type ConfigJSON struct {
	RuntimeOptions struct {
		Framework struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"framework"`
		ApplyPatches *bool `json:"applyPatches"`
	} `json:"runtimeOptions"`
}

func (d *FrameworkDependentDeployment) InstallFrameworks() error {
	var (
		aspnetcoreVersion, runtimeVersion string
		err                               error
	)
	if d.FrameworkName == "Microsoft.AspNetCore.All" || d.FrameworkName == "Microsoft.AspNetCore.App" {
		aspnetcoreVersion, err = d.Project.FindMatchingFrameworkVersion("dotnet-aspnetcore", d.FrameworkVersion, d.ApplyPatches)
		if err != nil {
			return err
		}
		aspnetConfigJSON, err := ParseRuntimeConfig(filepath.Join(d.DepDir, "dotnet-sdk", "shared", "Microsoft.AspNetCore.App", aspnetcoreVersion, "Microsoft.AspNetCore.App.runtimeconfig.json"))
		if err != nil {
			return err
		}
		runtimeVersion, err = d.Project.FindMatchingFrameworkVersion("dotnet-runtime", aspnetConfigJSON.RuntimeOptions.Framework.Version, d.ApplyPatches)
		if err != nil {
			return err
		}

	} else if d.FrameworkName == "Microsoft.NETCore.App" {
		runtimeVersion, err = d.Project.FindMatchingFrameworkVersion("dotnet-runtime", d.FrameworkVersion, d.ApplyPatches)
		if err != nil {
			return err
		}
		aspnetcoreVersion, err = d.Project.GetVersionFromDepsJSON()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalid framework specified in application runtime config file")
	}

	if err := d.Installer.InstallDependency(libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: aspnetcoreVersion}, filepath.Join(d.DepDir, "dotnet-sdk")); err != nil {
		return err
	}

	return d.Installer.InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: runtimeVersion}, filepath.Join(d.DepDir, "dotnet-sdk"))
}

func (d *SelfContainedDeployment) InstallFrameworks() error {
	return nil
}

func (d *SourceDeployment) InstallFrameworks() error {
	mainPath, err := d.Project.MainPath()
	if err != nil {
		return err
	}

	runtimeRegex := "<RuntimeFrameworkVersion>(*)</RuntimeFrameworkVersion>"
	aspnetcoreRegex := `"Microsoft.AspNetCore.All" Version="(.*)"`

	runtimeVersion, err := d.Project.VersionFromProjFile(mainPath, runtimeRegex, "dotnet-runtime")
	if err != nil {
		return err
	}
	aspnetcoreVersion, err := d.Project.VersionFromProjFile(mainPath, aspnetcoreRegex, "dotnet-aspnetcore")
	if err != nil {
		return err
	}

	if err := d.Installer.InstallDependency(libbuildpack.Dependency{Name: "dotnet-aspnetcore", Version: aspnetcoreVersion}, filepath.Join(d.DepDir, "dotnet-sdk")); err != nil {
		return err
	}
	if err := d.Installer.InstallDependency(libbuildpack.Dependency{Name: "dotnet-runtime", Version: runtimeVersion}, filepath.Join(d.DepDir, "dotnet-sdk")); err != nil {
		return err
	}

	return nil
}

func ParseRuntimeConfig(runtimeConfig string) (ConfigJSON, error) {
	obj := ConfigJSON{}
	if err := libbuildpack.NewJSON().Load(runtimeConfig, &obj); err != nil {
		return ConfigJSON{}, err
	}
	return obj, nil
}
