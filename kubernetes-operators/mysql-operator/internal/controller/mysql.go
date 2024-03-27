package controller

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	v1 "mysql-operator/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	persistentVolumeMode = corev1.PersistentVolumeFilesystem
	hostPathType         = corev1.HostPathType("")
)

func (r *MySQLReconciler) createPersistentVolume(ctx context.Context, mysql *v1.MySQL) error {
	l := log.FromContext(ctx)
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-pv", mysql.Name),
			Namespace: mysql.Namespace,
			Labels: map[string]string{
				"pv-usage": mysql.Name,
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(mysql.Spec.StorageSize),
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: fmt.Sprintf("/tmp/hostpath_pv/%s-pv/", mysql.Name),
					Type: &hostPathType,
				},
			},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			StorageClassName:              "standard",
			VolumeMode:                    &persistentVolumeMode,
		},
	}

	currPV := &corev1.PersistentVolume{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name: fmt.Sprintf("%s-pv", mysql.Name),
	}, currPV)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		l.Info(fmt.Sprintf("creating persistent volume '%s-pv'...", mysql.Name))
		return r.Client.Create(ctx, pv)
	}

	return nil
}

func (r *MySQLReconciler) createPersistentVolumeClaim(ctx context.Context, mysql *v1.MySQL) error {
	l := log.FromContext(ctx)
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-pvc", mysql.Name),
			Namespace: mysql.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(mysql.Spec.StorageSize),
				},
			},
			VolumeMode: &persistentVolumeMode,
			VolumeName: fmt.Sprintf("%s-pv", mysql.Name),
		},
	}

	err := ctrl.SetControllerReference(mysql, pvc, r.Scheme)
	if err != nil {
		return err
	}

	currPVC := &corev1.PersistentVolumeClaim{}
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("%s-pvc", mysql.Name),
		Namespace: mysql.Namespace,
	}, currPVC)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		l.Info(fmt.Sprintf("creating persistent volume claim '%s-pvc'...", mysql.Name))
		return r.Client.Create(ctx, pvc)
	}

	return nil
}

func (r *MySQLReconciler) createService(ctx context.Context, mysql *v1.MySQL) error {
	l := log.FromContext(ctx)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysql.Name,
			Namespace: mysql.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       int32(3306),
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{IntVal: int32(3306)},
				},
			},
			Selector: map[string]string{
				"app": mysql.Name,
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	err := ctrl.SetControllerReference(mysql, service, r.Scheme)
	if err != nil {
		return err
	}

	currService := &corev1.Service{}
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      mysql.Name,
		Namespace: mysql.Namespace,
	}, currService)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		l.Info(fmt.Sprintf("creating service '%s'...", mysql.Name))
		return r.Client.Create(ctx, service)
	}

	return nil
}

func (r *MySQLReconciler) createDeployment(ctx context.Context, mysql *v1.MySQL) error {
	l := log.FromContext(ctx)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysql.Name,
			Namespace: mysql.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": mysql.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": mysql.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: mysql.Name,
							Env: []corev1.EnvVar{
								{
									Name:  "MYSQL_ROOT_PASSWORD",
									Value: mysql.Spec.Password,
								},
								{
									Name:  "MYSQL_DATABASE",
									Value: mysql.Spec.Database,
								},
							},
							Image:           mysql.Spec.Image,
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									Name:          "mysql",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 3306,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"mysql",
											"-uroot",
											fmt.Sprintf("-p%s", mysql.Spec.Password),
											"-h",
											"127.0.0.1",
											"-e",
											"SELECT 1",
										},
									},
								},
								FailureThreshold:    int32(12),
								InitialDelaySeconds: int32(5),
								PeriodSeconds:       int32(5),
								SuccessThreshold:    int32(1),
								TimeoutSeconds:      int32(5),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/var/lib/mysql",
									Name:      "data",
								},
							},
						},
					},
					ServiceAccountName: "mysql-operator",
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: fmt.Sprintf("%s-pvc", mysql.Name),
								},
							},
						},
					},
				},
			},
		},
	}

	err := ctrl.SetControllerReference(mysql, deployment, r.Scheme)
	if err != nil {
		return err
	}

	currDeployment := &appsv1.Deployment{}
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      mysql.Name,
		Namespace: mysql.Namespace,
	}, currDeployment)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		l.Info(fmt.Sprintf("creating deployment '%s'...", mysql.Name))
		return r.Client.Create(ctx, deployment)
	}

	return nil
}
