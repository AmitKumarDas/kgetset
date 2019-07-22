### kgetset
kubernetes get and set

### Motivation - Test/Verify
#### Use Case 1
- Setup:
  - Apply a CRD into K8s
- Create corresponding CR into K8s
- Fetch this CR & do a deep equal check
- Teardown:
  - Delete the CRD from K8s
  - Verify no CR object exist

#### Use Case 2
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

### Assumtpions
- K8s Version 1.12
