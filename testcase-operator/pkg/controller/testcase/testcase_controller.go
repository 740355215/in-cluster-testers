package testcase

import (
	"cloud.tencent.com/tke/in-cluster-tester/testcase-operator/pkg/utils"
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"strconv"
	"time"

	tkev1 "cloud.tencent.com/tke/in-cluster-tester/testcase-operator/pkg/apis/tke/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	NormalExit = 0

	TestcasePassed = "passed"
	TestcaseFailed = "failed"

	PodNamespace = "default"

	ConfigMapName            = "testcase-configmap"
	ConfigMapNameInContainer = "testcase-conf"
	ConfigMapPathInContainer = "/etc/testcase/configmap"
)

var (
	log          = logf.Log.WithName("controller_testcase")
	SidecarImage = os.Getenv("SIDECAR_IMAGE")
	SidecarPort  = os.Getenv("SIDECAR_PORT")

	DelayPeriodAfterPassed = os.Getenv("DELAY_SECONDS_AFTER_TESETER_PASSED")
	DelayPeriodAfterFailed = os.Getenv("DELAY_SECONDS_AFTER_TESETER_FAILED")

	TestcaseSummary = os.Getenv("TESTCASE_SUMMARY")
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new TestCase Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTestCase{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("testcase-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return false
		},
	}
	// Watch for changes to primary resource TestCase
	err = c.Watch(&source.Kind{Type: &tkev1.TestCase{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner TestCase
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tkev1.TestCase{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileTestCase{}

// ReconcileTestCase reconciles a TestCase object
type ReconcileTestCase struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a TestCase object and makes changes based on the state read
// and what is in the TestCase.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTestCase) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	request.Namespace = ""
	request.NamespacedName.Namespace = ""
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name, "Request.NamespacedName", request.NamespacedName)
	reqLogger.Info("Reconciling TestCase")

	// Fetch the TestCase instance
	instance := &tkev1.TestCase{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if isCrdProceessed(instance) {
		return reconcile.Result{}, nil
	}

	// Define a new Pod object
	pod := newPodForCR(instance)
	pod.Namespace = PodNamespace

	// Set TestCase instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Pod exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name, "pod phase", found.Status.Phase)

	if utils.IsPodFailed(found) || utils.IsPodSucceeded(found) {
		if err := r.client.Delete(context.TODO(), found); err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	if utils.IsPodRunning(found) && isTestCaseContainerDone(pod, found) {
		if err := syncSecondResource(pod, found, instance, r); err != nil {
			return reconcile.Result{}, err
		}

		if err := syncSummaryStatus(pod.Namespace, r); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *tkev1.TestCase) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    cr.Name,
					Image:   cr.Spec.Image,
					Command: cr.Spec.Commands,
					EnvFrom: []corev1.EnvFromSource{
						{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: ConfigMapName},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      ConfigMapNameInContainer,
							MountPath: ConfigMapPathInContainer,
						},
					},
				},
				{
					Name:  "sidecar",
					Image: SidecarImage,
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
			Volumes: []corev1.Volume{
				{
					Name: ConfigMapNameInContainer,
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: ConfigMapName},
						},
					},
				},
			},
		},
	}
}

func syncSecondResource(pod, found *corev1.Pod, instance *tkev1.TestCase, r *ReconcileTestCase) error {
	log.Info("begin syncing second resource...")

	exitCode := getExitCode(pod, found)
	msg, err := getTesterMessageFromSidecar(found)
	if err != nil {
		return err
	}

	// update status to crd instance correspondingly
	instance.Status = tkev1.TestCaseStatus{
		Result:       getTestCaseResultByExitCode(exitCode),
		Message:      msg,
		CompleteTime: time.Now().Format("2006-01-02 15:04:05"),
		PodName:      found.Namespace + "/" + found.Name,
	}
	log.Info("begin updating crd status:", "instance", instance)

	if err := r.client.Update(context.TODO(), instance); err != nil {
		return err
	}

	// delay delete pod by close sidecar
	if err := delayDeletePod(found, exitCode); err != nil {
		return err
	}

	return nil
}

func isTestCaseContainerDone(pod, found *corev1.Pod) bool {
	for _, status := range found.Status.ContainerStatuses {
		if status.Name == pod.Spec.Containers[0].Name && status.State.Terminated != nil {
			return true
		}
	}

	return false
}

func getExitCode(pod, found *corev1.Pod) int32 {
	for _, status := range found.Status.ContainerStatuses {
		if status.Name == pod.Spec.Containers[0].Name && status.State.Terminated != nil {
			return status.State.Terminated.ExitCode
		}
	}
	return math.MinInt32
}

func getTesterMessageFromSidecar(pod *corev1.Pod) (string, error) {
	podIp := pod.Status.PodIP
	// get tester message
	msgUrl := "http://" + podIp + ":" + SidecarPort + "/message"
	log.Info(fmt.Sprintf("url is: %v\n", msgUrl))

	resp, err := http.Get(msgUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func delayDeletePod(pod *corev1.Pod, exitCode int32) error {
	sec, err := getDelaySecondsForDeletingPodByExitCode(exitCode)
	if err != nil {
		return err
	}

	podIp := pod.Status.PodIP
	//delayUrl := "http://" + podIp + ":" + SidecarPort + "/delay?seconds=" + string(sec)
	delayUrl := fmt.Sprintf("http://%v:%v/delay?seconds=%v", podIp, SidecarPort, sec)
	log.Info(fmt.Sprintf("delay url is: %v\n", delayUrl))

	log.Info(fmt.Sprintf("we will delete pod after %v secondes", sec))
	resp, err := http.Post(delayUrl, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func getTestCaseResultByExitCode(exitCode int32) string {
	// get tester result by exitcode(if exitcode==0, passed)
	if exitCode == NormalExit {
		return TestcasePassed
	}

	return TestcaseFailed
}

func getDelaySecondsForDeletingPodByExitCode(exitCode int32) (int, error) {
	delayTime := DelayPeriodAfterFailed
	if exitCode == NormalExit {
		delayTime = DelayPeriodAfterPassed
	}

	sec, err := strconv.Atoi(delayTime)
	if err != nil {
		log.Error(err, "DelayPeriodAfterFailed or DelayPeriodAfterPassed is not a number")
		return 0, err
	}

	return sec, nil
}

func syncSummaryStatus(ns string, r *ReconcileTestCase) error {

	summaryList := &tkev1.SummaryList{}
	if err := r.client.List(context.TODO(), nil, summaryList); err != nil {
		log.Info(fmt.Sprintf("list summary failed: %v", err))
		return err
	}
	summaryInstance := tkev1.Summary{}
	for _, summary := range summaryList.Items {
		if summary.Name == TestcaseSummary {
			summaryInstance = summary
			break
		}
	}

	//summaryInstance := &tkev1.Summary{}
	//if err := r.client.Get(context.TODO(), types.NamespacedName{Name: TestcaseSummary, Namespace: ns}, summaryInstance); err != nil {
	//	log.Info(fmt.Sprintf("get %v failed: %v", TestcaseSummary, err))
	//	return err
	//}

	tcList := &tkev1.TestCaseList{}
	if err := r.client.List(context.TODO(), nil, tcList); err != nil {
		log.Info(fmt.Sprintf("list testcase failed: %v", err))
		return err
	}

	summaryInstance.Status = getSummaryStatus(tcList)
	if err := r.client.Update(context.TODO(), &summaryInstance); err != nil {
		log.Info(fmt.Sprintf("update %v failed: %v", TestcaseSummary, err))
		return err
	}
	return nil
}

func getSummaryStatus(tcList *tkev1.TestCaseList) tkev1.SummaryStatus {
	total := len(tcList.Items)
	var passed, failed, testing int
	for _, tc := range tcList.Items {
		switch tc.Status.Result {
		case TestcasePassed:
			passed++
		case TestcaseFailed:
			failed++
		default:
			testing++
		}
	}

	log.Info(fmt.Sprintf("total:%v, pass:%v, fail:%v, testing:%v", total, passed, failed, testing))

	return tkev1.SummaryStatus{
		TotalNumber:  total,
		PassedNumber: passed,
		FailedNumber: failed,
	}
}

func isCrdProceessed(instance *tkev1.TestCase) bool {
	if instance.Status.Result != "" {
		return true
	}

	return false
}
