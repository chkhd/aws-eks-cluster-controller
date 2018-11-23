package eks

import (
	"testing"
	"time"

	clusterv1alpha1 "github.com/awslabs/aws-eks-cluster-controller/pkg/apis/cluster/v1alpha1"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}
var controlPlaneKey = types.NamespacedName{Name: "foo-controlplane", Namespace: "default"}
var nodeGroup1Key = types.NamespacedName{Name: "foo-nodegroup-1", Namespace: "default"}
var nodeGroup2Key = types.NamespacedName{Name: "foo-nodegroup-2", Namespace: "default"}

const timeout = time.Second * 25

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	instance := &clusterv1alpha1.EKS{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
		Spec: clusterv1alpha1.EKSSpec{
			AccountID: "1234foo",
			ControlPlane: clusterv1alpha1.ControlPlaneSpec{
				ClusterName: "cluster-stuff",
				StackName:   "stack-stuff",
			},
			NodeGroups: []clusterv1alpha1.NodeGroupSpec{
				clusterv1alpha1.NodeGroupSpec{},
				clusterv1alpha1.NodeGroupSpec{},
			},
		},
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	recFn, requests := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create the EKS object and expect the Reconcile and Deployment to be created
	err = c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	controlPlane := &clusterv1alpha1.ControlPlane{}
	g.Eventually(func() error { return c.Get(context.TODO(), controlPlaneKey, controlPlane) }, timeout).
		Should(gomega.Succeed())

	// controlPlane.Status.Status = "Complete"
	// g.Expect(c.Update(context.TODO(), controlPlane)).Should(gomega.Succeed())

	// nodeGroup1 := &clusterv1alpha1.NodeGroup{}
	// g.Eventually(func() error { return c.Get(context.TODO(), nodeGroup1Key, nodeGroup1) }, timeout).
	// 	Should(gomega.Succeed())

	// nodeGroup2 := &clusterv1alpha1.NodeGroup{}
	// g.Eventually(func() error { return c.Get(context.TODO(), nodeGroup2Key, nodeGroup2) }, timeout).
	// 	Should(gomega.Succeed())

}