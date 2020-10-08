module github.com/pulumi/kubespy

go 1.15

require (
	github.com/fatih/color v1.9.0
	github.com/gosuri/uilive v0.0.0-20170323041506-ac356e6e42cd // indirect
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/mbrlabs/uilive v0.0.0-20170420192653-e481c8e66f15
	github.com/pulumi/pulumi-kubernetes/provider/v2 v2.0.1-0.20201007200217-2c206f417da4
	github.com/spf13/cobra v1.0.0
	github.com/yudai/gojsondiff v0.0.0-20170107030110-7b1b7adf999d
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	github.com/yudai/pp v2.0.1+incompatible // indirect
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.1+incompatible
	github.com/evanphx/json-patch => github.com/evanphx/json-patch v0.0.0-20200808040245-162e5629780b // 162e5629780b is the SHA for git tag v4.8.0
)
