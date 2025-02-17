/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"sigs.k8s.io/azurelustre-csi-driver/pkg/azurelustre"

	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"
)

func init() {
	_ = flag.Set("logtostderr", "true")
}

var (
	endpoint                   = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID                     = flag.String("nodeid", "", "node id")
	version                    = flag.Bool("version", false, "Print the version and exit.")
	metricsAddress             = flag.String("metrics-address", "0.0.0.0:29634", "export the metrics")
	driverName                 = flag.String("drivername", azurelustre.DefaultDriverName, "name of the driver")
	enableAzureLustreMockMount = flag.Bool("enable-azurelustre-mock-mount", false, "Whether enable mock mount(only for testing)")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	if *version {
		info, err := azurelustre.GetVersionYAML(*driverName)
		if err != nil {
			klog.Fatalln(err)
		}
		fmt.Println(info)
		os.Exit(0)
	}

	exportMetrics()
	handle()
	os.Exit(0)
}

func handle() {
	driverOptions := azurelustre.DriverOptions{
		NodeID:                     *nodeID,
		DriverName:                 *driverName,
		EnableAzureLustreMockMount: *enableAzureLustreMockMount,
	}
	driver := azurelustre.NewDriver(&driverOptions)
	if driver == nil {
		klog.Fatalln("Failed to initialize Azure Lustre CSI driver")
	}
	driver.Run(*endpoint, "", false)
}

func exportMetrics() {
	l, err := net.Listen("tcp", *metricsAddress)
	if err != nil {
		klog.Warningf("failed to get listener for metrics endpoint: %v", err)
		return
	}
	serve(context.Background(), l, serveMetrics)
}

func serve(ctx context.Context, l net.Listener, serveFunc func(net.Listener) error) {
	path := l.Addr().String()
	klog.V(2).Infof("set up prometheus server on %v", path)
	go func() {
		defer l.Close()
		if err := serveFunc(l); err != nil {
			klog.Fatalf("serve failure(%v), address(%v)", err, path)
		}
	}()
}

func serveMetrics(l net.Listener) error {
	m := http.NewServeMux()

	// azure cloud provider uses legacyregistry currently
	m.Handle("/metrics", legacyregistry.Handler()) //nolint
	return trapClosedConnErr(http.Serve(l, m))
}

func trapClosedConnErr(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "use of closed network connection") {
		return nil
	}
	return err
}
