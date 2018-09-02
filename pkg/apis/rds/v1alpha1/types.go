package v1alpha1

import (
	"crypto/rand"
	"encoding/hex"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// State represents the state.
const (
	StatePending = "Pending"
	StateCreated = "Created"
	StateFailure = "Failure"
)

// DatabaseList lists the database.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Database `json:"items"`
}

// Database object creates RDS databases.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DatabaseSpec   `json:"spec"`
	Status            DatabaseStatus `json:"status,omitempty"`
}

// DatabaseSpec configures the RDS database.
type DatabaseSpec struct {
	Engine                  string   `json:"engine"`
	EngineVersion           string   `json:"engineVersion"`
	Username                string   `json:"username"`
	Password                string   `json:"password"`
	Database                string   `json:"database"`
	Storage                 int64    `json:"storage"`
	AutoMinorVersionUpgrade bool     `json:"autoMinorVersionUpgrade"`
	AvailabilityZone        string   `json:"availabilityZone"`
	BackupRetentionPeriod   int64    `json:"backupRetentionPeriod"`
	CharacterSetName        string   `json:"characterSetName"`
	InstanceClass           string   `json:"instanceClass"`
	SubnetGroup             string   `json:"subnetGroup"`
	Iops                    int64    `json:"iops"`
	MultiAZ                 bool     `json:"multiAz"`
	Encrypted               bool     `json:"encrypted"`
	StorageType             string   `json:"storageType"`
	SecurityGroups          []string `json:"securityGroups"`
}

// Defaults will set default configuration.
func Defaults(db *Database) {
	s := db.Spec
	if s.Engine == "" {
		s.Engine = "postgres"
	}
	if s.EngineVersion == "" {
		s.EngineVersion = "10.4"
	}
	if s.Username == "" {
		s.Username = "postgres"
	}
	if s.Password == "" {
		b := make([]byte, 32)
		rand.Read(b)
		s.Password = hex.EncodeToString(b)
	}
	if s.Database == "" {
		s.Database = "postgres"
	}
	if s.StorageType == "" {
		s.StorageType = "gp2"
	}
	if s.InstanceClass == "" {
		s.InstanceClass = "db.t2.micro"
	}
	if s.Storage == 0 {
		s.Storage = 20
	}
}

// DatabaseStatus holds state and error structs.
type DatabaseStatus struct {
	State string `json:"state"`
	Error string `json:"error"`
}
