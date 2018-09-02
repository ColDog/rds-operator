# RDS Operator

## Example

```yaml
apiVersion: "rds.aws.com/v1alpha1"
kind: "Database"
metadata:
  name: "example"
spec:
  engine: postgres
  engineVersion: "10.4"
  username: postgres
  password: <nil>
  database: postgres
  storage: 20
  storageType: gp2
  autoMinorVersionUpgrade: false
  availabilityZone: us-west-2
  backupRetentionPeriod: 7
  characterSetName: utf8
  instanceClass: db.t2.micro
  subnetGroup: <subnet>
  multiAz: false
  encrypted: false
  dbSecurityGroups: []
  vpcSecurityGroups: []
```
