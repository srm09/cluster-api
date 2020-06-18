/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/klog"
	"k8s.io/klog/klogr"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	"sigs.k8s.io/cluster-api/cmd/clusterctl/log"
	"sigs.k8s.io/cluster-api/test/helpers"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func init() {
	klog.InitFlags(nil)
	log.SetLogger(klogr.New())
}

const (
	timeout = time.Second * 10
)

var (
	testEnv           *helpers.TestEnvironment
	clusterReconciler *ClusterReconciler
	ctx               = context.Background()
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	By("bootstrapping test environment")
	var err error
	testEnv, err = helpers.NewTestEnvironment()
	Expect(err).NotTo(HaveOccurred())

	clusterReconciler = &ClusterReconciler{
		Client:   testEnv,
		Log:      log.Log,
		recorder: testEnv.GetEventRecorderFor("cluster-controller"),
	}
	Expect(clusterReconciler.SetupWithManager(testEnv.Manager, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())
	Expect((&MachineReconciler{
		Client:   testEnv,
		Log:      log.Log,
		recorder: testEnv.GetEventRecorderFor("machine-controller"),
	}).SetupWithManager(testEnv.Manager, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())
	Expect((&MachineSetReconciler{
		Client:   testEnv,
		Log:      log.Log,
		recorder: testEnv.GetEventRecorderFor("machineset-controller"),
	}).SetupWithManager(testEnv.Manager, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())
	Expect((&MachineDeploymentReconciler{
		Client:   testEnv,
		Log:      log.Log,
		recorder: testEnv.GetEventRecorderFor("machinedeployment-controller"),
	}).SetupWithManager(testEnv.Manager, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())
	Expect((&MachineHealthCheckReconciler{
		Client:   testEnv,
		Log:      log.Log,
		recorder: testEnv.GetEventRecorderFor("machinehealthcheck-controller"),
	}).SetupWithManager(testEnv.Manager, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	By("starting the manager")
	go func() {
		Expect(testEnv.StartManager()).To(Succeed())
	}()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	if testEnv != nil {
		By("tearing down the test environment")
		Expect(testEnv.Stop()).To(Succeed())
	}
})

func ContainRefOfKind(kind string) types.GomegaMatcher {
	return &refKindMatcher{
		kind: kind,
	}
}

type refKindMatcher struct {
	kind string
}

func (matcher *refKindMatcher) Match(actual interface{}) (success bool, err error) {
	ownerRefs, ok := actual.([]metav1.OwnerReference)
	if !ok {
		return false, fmt.Errorf("ContainRefOfKind matcher expects a slice of OwnerReference")
	}

	for _, ref := range ownerRefs {
		if ref.Kind == matcher.kind {
			return true, nil
		}
	}

	return false, nil
}

func (matcher *refKindMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %+v to contain refs of Kind %s", actual, matcher.kind)
}

func (matcher *refKindMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %+v not to contain refs of Kind %s", actual, matcher.kind)
}
