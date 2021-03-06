package sysd

import (
	"fmt"
	"strings"
)

type conditional interface {
	conditionKey() string
	conditionArg() string
}

// Conditions describes a set of checks which must pass for the unit to run.
type Conditions []conditional

// String returns the conditions serialized into config lines.
func (c *Conditions) String() string {
	var out strings.Builder
	for _, cond := range *c {
		out.WriteString(fmt.Sprintf("%s=%s\n", cond.conditionKey(), cond.conditionArg()))
	}
	return out.String()
}

// ConditionExists specifies a file must exist at the path.
type ConditionExists string

func (c ConditionExists) conditionKey() string {
	return "ConditionPathExists"
}
func (c ConditionExists) conditionArg() string {
	return string(c)
}
func (c ConditionExists) String() string {
	return "Exists(" + string(c) + ")"
}

// ConditionNotExists specifies a file must be missing at the path.
type ConditionNotExists string

func (c ConditionNotExists) conditionKey() string {
	return "ConditionPathNotExists"
}
func (c ConditionNotExists) conditionArg() string {
	return string(c)
}
func (c ConditionNotExists) String() string {
	return "NotExists(" + string(c) + ")"
}

// ConditionHost specifies the machine must be a given hostname or machine ID.
type ConditionHost string

func (c ConditionHost) conditionKey() string {
	return "ConditionHost"
}
func (c ConditionHost) conditionArg() string {
	return string(c)
}
func (c ConditionHost) String() string {
	return "Host(" + string(c) + ")"
}

// ConditionFirstBoot specifies the current boot must be the first.
type ConditionFirstBoot string

func (c ConditionFirstBoot) conditionKey() string {
	return "ConditionFirstBoot"
}
func (c ConditionFirstBoot) conditionArg() string {
	return string(c)
}
func (c ConditionFirstBoot) String() string {
	return "ConditionFirstBoot(" + string(c) + ")"
}
