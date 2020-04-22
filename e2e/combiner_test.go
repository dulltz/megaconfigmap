package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Combiner", func() {
	Describe("Creating MegaConfigMap", func() {
		Context("With 1MB file", func() {
			It("should be combined into one file", func() {
				By("preparing file")
				dummyFile := "data.dummy"
				f, err := os.Create(dummyFile)
				Expect(err).ShouldNot(HaveOccurred())
				defer os.Remove(dummyFile)
				Expect(f.Truncate(1e6)).ShouldNot(HaveOccurred())
				Expect(f.Close()).ShouldNot(HaveOccurred())
				stdout, stderr, err := run("kubectl", "megaconfigmap", "create", "my-conf", "--from-file=./"+dummyFile)
				Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s", stdout.String(), stderr.String())
				By("creating a pod with combiner")
				stdout, stderr, err = run("kubectl", "apply", "-f", "../examples/pod.yaml")
				Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s", stdout.String(), stderr.String())
				Eventually(func() error {
					stdout, stderr, err := run("kubectl", "exec", "megaconfigmap-demo", "--", "ls", "-lh", "/demo/"+dummyFile)
					if err != nil {
						return fmt.Errorf("err: %s, stdout: %s, stderr: %s", err, stdout.String(), stderr.String())
					}
					return nil
				}, 60*time.Second).ShouldNot(HaveOccurred())

				By("delete all partial configmaps if parent have been deleted")
				stdout, stderr, err = run("kubectl", "delete", "cm", "my-conf")
				Eventually(func() error {
					stdout, stderr, err := run("kubectl", "get", "cm", "-o=json")
					if err != nil {
						return fmt.Errorf("err: %s, stdout: %s, stderr: %s", err, stdout.String(), stderr.String())
					}
					var cml corev1.ConfigMapList
					err = json.Unmarshal(stdout.Bytes(), &cml)
					if err != nil {
						return fmt.Errorf("failed to unmarshal. err: %s", err)
					}
					if len(cml.Items) > 0 {
						return fmt.Errorf(string(len(cml.Items)) + " configmap remains")
					}
					return nil
				}, 20*time.Second).ShouldNot(HaveOccurred())

				By("clean up")
				stdout, stderr, err = run("kubectl", "delete", "-f", "../examples/pod.yaml")
				Expect(err).ShouldNot(HaveOccurred(), "stdout: %s, stderr: %s", stdout.String(), stderr.String())
			})
		})
	})
})

func run(first string, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	cmd := exec.Command(first, args...)
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf
	err := cmd.Run()
	return outBuf, errBuf, err
}
