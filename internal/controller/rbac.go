// +kubebuilder:rbac:groups=stalwart.toando.de,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=stalwart.toando.de,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=stalwart.toando.de,resources=clusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
package controller
