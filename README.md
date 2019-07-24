### kgetset
kgetset is a verification tool to test CustomResourceDefinition(s) and CustomResource(s)

### Testsuite
#### Test Case 1 [WIP]
- Bring up a microk8s cluster
- Run `microk8s.kubectl apply -f crd.yaml`
  - This job does below:

  ```
  - Setup:
    - Apply a CRD
  - When:
    - Create a CR at K8s
    - Fetch this CR from K8s
  - Then:
    - expected CR == actual CR
  - Teardown:
    - Delete the CRD
  ```


#### Test Case 2 [Not Started]
- Setup:
  - Apply a CRD into K8s
- Create corresponding CR into K8s
- Design a new schema:
  - use same CRD
  - use same GVK
  - Map values from old CR 
  - Set new values for new fields
- Create this new schema into K8s
- Fetch this new schema from K8s & do a deep equal check
- Fetch the old schema from K8s & do a deep equal check
- Teardown:
  - Delete the CRD from k8s
  - Verify no CR objects exist
