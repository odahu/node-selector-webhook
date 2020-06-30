This repository contains webhook for kubernetes that assign 
a node selector to pods in specific namespace

### How it works

Webhook monitors all pods in namespaces that 
labeled by "odahu/node-selector-webhook" (label should exist)
and adds configured tolerations and node selector for all pods 