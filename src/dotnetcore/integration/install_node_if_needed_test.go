package integration_test

import (
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/agouti"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Dotnet Buildpack", func() {
	var app *cutlass.App
	var page *agouti.Page

	BeforeEach(func() {
		var err error
		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		PrintFailureLogs(app.Name)
		app = DestroyApp(app)
		Expect(page.Destroy()).To(Succeed())
	})

	Context("Deploying a react app", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "react_node_app"))
			app.Disk = "2G"
			app.Memory = "2G"
		})

		It("displays homepage when install_node set to true", func() {
			app.SetEnv("INSTALL_NODE", "true")
			//TODO: Add expect to ensure that install_node is actually set to true
			PushAppAndConfirm(app)

			url, err := app.GetUrl("/")
			Expect(err).NotTo(HaveOccurred())

			Expect(page.Navigate(url)).To(Succeed())
			Expect(app.Stdout).To(ContainSubstring("Keeping Node"))
			Eventually(page.HTML, 10*time.Second).Should(ContainSubstring("Hello, world!"))
		})

		FIt("displays homepage when install_node is not set", func() {
			PushAppAndConfirm(app)

			url, err := app.GetUrl("/")
			Expect(err).NotTo(HaveOccurred())

			Expect(page.Navigate(url)).To(Succeed())
			Expect(app.Stdout).To(ContainSubstring("Keeping Node"))
			Eventually(page.HTML, 10*time.Second).Should(ContainSubstring("Hello, world!"))
		})
	})
})
