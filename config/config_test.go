package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"testing"

	k8sInfra "github.com/metrue/fx/infra/k8s"
	"github.com/metrue/fx/types"
)

func TestConfig(t *testing.T) {
	configPath := "./tmp/config.yml"
	defer func() {
		if err := os.RemoveAll("./tmp/config.yml"); err != nil {
			t.Fatal(err)
		}
	}()

	// default cloud
	c, err := Load(configPath)
	if err != nil {
		t.Fatal(err)
	}

	defaultMeta, err := c.GetCurrentCloud()
	if err != nil {
		t.Fatal(err)
	}
	var cloudMeta map[string]string
	if err := json.Unmarshal(defaultMeta, &cloudMeta); err != nil {
		t.Fatal(err)
	}
	if cloudMeta["ip"] != "127.0.0.1" {
		t.Fatalf("should get %s but got %s", "127.0.0.1", cloudMeta["ip"])
	}

	me, _ := user.Current()
	if cloudMeta["user"] != me.Username {
		t.Fatalf("should get %s but got %s", me.Username, cloudMeta["user"])
	}
	if cloudMeta["type"] != types.CloudTypeDocker {
		t.Fatalf("should get %s but got %s", types.CloudTypeDocker, cloudMeta["type"])
	}
	if cloudMeta["name"] != "default" {
		t.Fatalf("should get %s but got %s", "default", cloudMeta["name"])
	}

	n1, err := k8sInfra.CreateNode(
		"1.1.1.1",
		"user-1",
		"k3s-master",
		"master-node",
	)
	if err != nil {
		t.Fatal(err)
	}
	n2, err := k8sInfra.CreateNode(
		"1.1.1.1",
		"user-1",
		"k3s-agent",
		"agent-node-1",
	)
	if err != nil {
		t.Fatal(err)
	}

	kName := "k8s-1"
	kubeconf := "./tmp/" + kName + "config.yml"
	defer func() {
		if err := os.RemoveAll(kubeconf); err != nil {
			t.Fatal(err)
		}
	}()

	// add k8s cloud
	kCloud := k8sInfra.NewCloud(kubeconf, n1, n2)
	kMeta, err := kCloud.Dump()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.AddCloud(kName, kMeta); err != nil {
		t.Fatal(err)
	}

	curMeta, err := c.GetCurrentCloud()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(curMeta, defaultMeta) {
		t.Fatalf("should get %v but got %v", defaultMeta, curMeta)
	}

	if err := c.UseCloud("cloud-not-existed"); err == nil {
		t.Fatalf("should get error when there is not given cloud name")
	}

	if err := c.UseCloud(kName); err != nil {
		t.Fatal(err)
	}

	curMeta, err = c.GetCurrentCloud()
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(curMeta, kMeta) {
		t.Fatalf("should get %v but got %v", kMeta, curMeta)
	}

	body, err := c.View()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(body))

	dir, err := c.Dir()
	if err != nil {
		t.Fatal(err)
	}
	here, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if dir != filepath.Join(here, "./tmp") {
		t.Fatalf("should get %s but got %s", "./tmp", dir)
	}
}
