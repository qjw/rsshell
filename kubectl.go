package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

const (
	getNamespacesP  = "get namespaces --output=json"
	getPodsP        = "get pods -n %s --output=json"
	getDeploymentsP = "get deploy -n %s --output=json"
)

func kubectl(arg string, obj interface{}) error {
	args := strings.Split(arg, " ")
	out, err := exec.Command("kubectl", args...).Output()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	if err := json.Unmarshal(out, obj); err != nil {
		fmt.Println(err.Error())
		return err
	}

	// for debug
	//	out, err = json.MarshalIndent(obj, "\t\t\t\t", "  ")
	//	fmt.Println(string(out))
	// end debug
	return nil
}

type Namespace struct {
	ApiVersion string `json:"apiVersion,omitempty"`
	Items      []struct {
		Kind     string `json:"kind,omitempty"`
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
	} `json:"items"`
}

func getNamespaces() (*Namespace, error) {
	var obj Namespace
	if err := kubectl(getNamespacesP, &obj); err != nil {
		fmt.Printf(err.Error())
		return nil, err
	}

	return &obj, nil
}

type Env struct {
	Name      string `json:"name"`
	Value     string `json:"value,omitempty"`
	ValueFrom *struct {
		FieldRef *struct {
			ApiVersion string `json:"apiVersion,omitempty"`
			FieldPath  string `json:"fieldPath,omitempty"`
		} `json:"fieldRef,omitempty"`
		SecretKeyRef *struct {
			Key  string `json:"key"`
			Name string `json:"name"`
		} `json:"secretKeyRef,omitempty"`
	} `json:"valueFrom,omitempty"`
}

type Volume struct {
	Name     string `json:"name"`
	EmptyDir *struct {
		SizeLimit string `json:"sizeLimit,omitempty"`
	} `json:"emptyDir,omitempty"`
	ConfigMap *struct {
		DefaultMode int    `json:"defaultMode,omitempty"`
		Name        string `json:"name,omitempty"`
	} `json:"configMap,omitempty"`
	Secret *struct {
		DefaultMode int    `json:"defaultMode,omitempty"`
		SecretName  string `json:"secretName,omitempty"`
	} `json:"secret,omitempty"`
}

type Container struct {
	Name            string `json:"name"`
	Image           string `json:"image"`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
	NodeName        string `json:"nodeName,omitempty"`
	ServiceAccount  string `json:"serviceAccount,omitempty"`
	RestartPolicy   string `json:"restartPolicy,omitempty"`
	Env             []Env  `json:"env,omitempty"`
	Ports           []struct {
		Name          string `json:"name,omitempty"`
		ContainerPort int    `json:"containerPort,omitempty"`
		Protocol      string `json:"protocol,omitempty"`
	} `json:"ports,omitempty"`
	VolumeMounts []struct {
		Name      string `json:"name,omitempty"`
		MountPath string `json:"mountPath,omitempty"`
		ReadOnly  bool   `json:"readOnly,omitempty"`
	} `json:"volumeMounts,omitempty"`
}

type ContainerSpec struct {
	Containers     []Container `json:"containers"`
	Volumes        []Volume    `json:"volumes,omitempty"`
	ServiceAccount string      `json:"serviceAccount,omitempty"`
	SchedulerName  string      `json:"schedulerName,omitempty"`
	RestartPolicy  string      `json:"restartPolicy,omitempty"`
	DnsPolicy      string      `json:"dnsPolicy,omitempty"`
}

type Pod struct {
	ApiVersion string `json:"apiVersion,omitempty"`
	Items      []struct {
		Kind     string `json:"kind,omitempty"`
		Metadata struct {
			Name      string            `json:"name"`
			NameSpace string            `json:"namespace"`
			Labels    map[string]string `json:"labels"`
		} `json:"metadata"`
		Spec ContainerSpec `json:"spec"`
	} `json:"items"`
}

func getPods(namespace string) (*Pod, error) {
	var obj Pod
	if err := kubectl(fmt.Sprintf(getPodsP, namespace), &obj); err != nil {
		fmt.Printf(err.Error())
		return nil, err
	}

	return &obj, nil
}

type Deployment struct {
	ApiVersion string `json:"apiVersion,omitempty"`
	Items      []struct {
		Kind     string `json:"kind,omitempty"`
		Metadata struct {
			Name      string            `json:"name"`
			NameSpace string            `json:"namespace"`
			Labels    map[string]string `json:"labels"`
		} `json:"metadata"`
		Spec struct {
			Replicas int `json:"replicas"`
			Template struct {
				Spec ContainerSpec `json:"spec"`
			} `json:"template"`
		} `json:"spec"`
	} `json:"items"`
}

func getDeployments(namespace string) (*Deployment, error) {
	var obj Deployment
	if err := kubectl(fmt.Sprintf(getDeploymentsP, namespace), &obj); err != nil {
		fmt.Printf(err.Error())
		return nil, err
	}

	return &obj, nil
}

func check(obj interface{}, err error) {
	if err != nil {
		panic(err)
	}
	if out, err := json.MarshalIndent(obj, "", "  "); err == nil {
		fmt.Println(string(out))
	} else {
		panic(err)
	}
}

func main() {
	check(getNamespaces())
	check(getPods("tt"))
	check(getDeployments("tt"))
}

