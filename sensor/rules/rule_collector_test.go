package rules

import (
	"context"
	"testing"

	"github.com/luqmanMohammed/eventsrunner-k8s-sensor/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	exampleBasicRuleStr = `
	[{
		id": "basic-pod-rule",
		"group": "",
		"version": "v1",
		"resource": "pods",
		"namespaces": ["default"]
	}]
	`
	exampleBasicRule2Str = `
	[{
		"id": "basic-configmap-rule",
		"group": "",
		"version": "v1",
		"resource": "configmaps",
		"namespaces": ["default"]
	},{
		"id": "basic-namespace-rule",
		"group": "",
		"version": "v1",
		"resource": "namespaces"
	},{
		"id": "basic-pod-rule",
		"group": "",
		"version": "v1",
		"resource": "pods",
		"namespaces": ["default"],
		"eventTypes": ["ADDED", "MODIFIED"]
	}]
	`
)

func setupRuleCollector() *ConfigMapRuleCollector {
	config := utils.GetKubeAPIConfigOrDie("")
	return &ConfigMapRuleCollector{
		clientSet:                kubernetes.NewForConfigOrDie(config),
		sensorNamespace:          "default",
		sensorRuleConfigMapLabel: "er-sensor-rules",
	}
}

func addRuleConfigMap(configMapName string, strRule string) error {
	config := utils.GetKubeAPIConfigOrDie("")
	if _, err := kubernetes.NewForConfigOrDie(config).CoreV1().ConfigMaps("default").Create(context.Background(), &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configMapName,
			Labels: map[string]string{
				"er-sensor-rules": "true",
			},
		},
		Data: map[string]string{
			"rules": strRule,
		},
	}, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

func deleteRuleConfigMap(configMapName string) error {
	config := utils.GetKubeAPIConfigOrDie("")
	if err := kubernetes.NewForConfigOrDie(config).CoreV1().ConfigMaps("default").Delete(context.Background(), configMapName, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func TestStarterRuleCollectionFromMultipleConfigMaps(t *testing.T) {
	addRuleConfigMap("basic-rules", exampleBasicRuleStr)
	defer deleteRuleConfigMap("basic-rules")
	addRuleConfigMap("basic-rules2", exampleBasicRule2Str)
	defer deleteRuleConfigMap("basic-rules2")
	ruleCollector := setupRuleCollector()
	if collectedRules, err := ruleCollector.Collect(context.Background()); err != nil {
		t.Errorf("Error while collecting rules: %v", err)
	} else {
		if len(collectedRules) != 3 {
			t.Errorf("Expected 3 rules, got %d", len(collectedRules))
		}
		if collectedRules["basic-pod-rule"].EventTypes[0] != ADDED {
			t.Errorf("Expected event type %s, got %s", ADDED, collectedRules["basic-pod-rule"].EventTypes[0])
		}
	}
}
