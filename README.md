### kgetset
kgetset is a verification tool to test CustomResourceDefinition(s) and CustomResource(s)

### Testsuite
- Bring up a microk8s cluster
- Run `microk8s.kubectl apply -f suite.yaml`
  - This job verifies below test cases:

#### Test Case 1
- This is implemented in `hello` package
```
- Setup:
  - Apply a CRD into K8s
  - Fetch this CRD from K8s
  - Verify if both instances match
- Teardown:
  - Delete the CRD from K8s
```


#### Test Case 2 [WIP]
```
- Setup:
  - Apply a CRD into K8s
- When:
  - Create a CR of above CRD at K8s
  - Create a second CR of above CRD but using a new schema
- Then:
  - Fetch first CR from K8s & match with local instance
  - Fetch second CR from k8s & match with local instance
- Teardown:
  - Delete the CRD from K8s
- PostTeardown:
  - Verify if all the CRs get deleted
```
