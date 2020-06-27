This repository contains webhook for kubernetes that assign 
a node selector to pods in specific namespace

### How it works

Webhook monitor all pods in namespaces that 
labeled by "odahu/node-selector-webhook" (label should exists)
and add configured tolerations and node selector for all pods 