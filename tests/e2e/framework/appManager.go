// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package framework

import (
	"path/filepath"

	"github.com/golang/glog"

	"istio.io/istio/tests/e2e/util"
)

const (
	kubeInjectPrefix = "KubeInject"
	hopYamlTmpl      = "tests/e2e/framework/testdata/hop.yam.tmpl"
)

// App gathers information for Hop app
type App struct {
	AppYamlTemplate string
	AppYaml         string
	KubeInject      bool
}

// AppManager organize and deploy apps
type AppManager struct {
	Apps      []*App
	tmpDir    string
	namespace string
	istioctl  *Istioctl
}

// Hop gathers information for Hop app
type Hop struct {
	*App
	AppImage   string
	Deployment string
	Service    string
	HTTPPort   int
	GRPCPort   int
	Version    string
}

// NewHop instantiate a Hop App based on the hopImage flag.
func NewHop(d, s, v string, h, g int) *Hop {
	return &Hop{
		App: &App{
			AppYamlTemplate: util.GetResourcePath(hopYamlTmpl),
		},
		Deployment: d,
		Service:    d,
		Version:    v,
		HTTPPort:   h,
		GRPCPort:   g,
	}
}

// NewAppManager create a new AppManager
func NewAppManager(tmpDir, namespace string, istioctl *Istioctl) *AppManager {
	return &AppManager{
		namespace: namespace,
		tmpDir:    tmpDir,
		istioctl:  istioctl,
	}
}

// generateAppYaml deploy testing app from tmpl
func (am *AppManager) generateAppYaml(a *App) error {
	if a.AppYamlTemplate == "" {
		return nil
	}
	var err error
	a.AppYaml, err = util.CreateTempfile(am.tmpDir, filepath.Base(a.AppYamlTemplate), yamlSuffix)
	if err != nil {
		return err
	}
	if err := util.Fill(a.AppYaml, a.AppYamlTemplate, a); err != nil {
		glog.Errorf("Failed to generate yaml for template %s", a.AppYamlTemplate)
		return err
	}
	return nil
}

func (am *AppManager) deploy(a *App) error {
	if err := am.generateAppYaml(a); err != nil {
		return err
	}
	finalYaml := a.AppYaml
	if a.KubeInject {
		var err error
		finalYaml, err = util.CreateTempfile(am.tmpDir, kubeInjectPrefix, yamlSuffix)
		if err != nil {
			return err
		}
		if err = am.istioctl.KubeInject(a.AppYaml, finalYaml); err != nil {
			return err
		}
	}
	if err := util.KubeApply(am.namespace, finalYaml); err != nil {
		glog.Errorf("Kubectl apply %s failed", finalYaml)
		return err
	}
	return nil
}

// Setup deploy apps
func (am *AppManager) Setup() error {
	glog.Info("Setting up apps")
	for _, a := range am.Apps {
		if err := am.deploy(a); err != nil {
			return err
		}
	}
	return nil
}

// Teardown currently does nothing, only to satisfied cleanable{}
func (am *AppManager) Teardown() error {
	return nil
}

// AddApp for automated deployment. Must be done before Setup Call.
func (am *AppManager) AddApp(a *App) {
	am.Apps = append(am.Apps, a)
}
