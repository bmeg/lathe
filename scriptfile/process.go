package scriptfile

type ProcessDesc struct {
	BasePath    string
	Name        string
	Desc        map[string]any
	CommandLine string
	Shell       string
	Inputs      map[string]string
	Outputs     map[string]string
	MemMB       uint
	NCpus       uint
}

func (pd *ProcessDesc) GetName() string {
	return pd.Name
}

func (pd *ProcessDesc) GetBasePath() string {
	return pd.BasePath
}

func (pd *ProcessDesc) GetInputs() map[string]string {
	return pd.Inputs
}

func (pd *ProcessDesc) GetProcess() *ProcessDesc {
	return pd
}
