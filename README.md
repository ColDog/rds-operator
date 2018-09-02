# RDS Operator


## Usage

Install the operator:

```bash
helm template --name rds-operator --namespace kube-system ./charts/rds-operator | kubectl create -f -
```

Check on the operator status:

```bash
kubectl -n kube-system get pods | grep rds-operator
```

Apply a database:

```bash
cat <<EOF | kubectl create -f -
apiVersion: "rds.aws.com/v1alpha1"
kind: "Database"
metadata:
  name: "example"
  namespace: "default"
spec:
  engine: postgres
  engineVersion: "10.4"
EOF
```

Fetch the database:

```bash
kubectl get databases -o yaml
```

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
