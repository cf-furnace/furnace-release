package models

type ModificationTag struct {
	Epoch string
	Index uint32
}

func (m *ModificationTag) SucceededBy(other *ModificationTag) bool {
	if m == nil || m.Epoch == "" || other.Epoch == "" {
		return true
	}

	return m.Epoch != other.Epoch || m.Index < other.Index
}

func (m *ModificationTag) Equal(other *ModificationTag) bool {
	if m == nil || other == nil {
		return false
	}
	return m.Epoch == other.Epoch && m.Index == other.Index
}
