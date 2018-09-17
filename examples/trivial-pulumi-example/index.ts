// Copyright 2016-2018, Pulumi Corporation.  All rights reserved.

import * as k8s from "@pulumi/kubernetes";

// Create an nginx pod
let nginx = new k8s.yaml.ConfigFile("yaml/nginx.yaml");
