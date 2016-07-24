package runner

type Docker struct {
	Image 		string
	Name 		string
	Ports 		[]string
	Net 		string
	NetAlias	string
	Env 		[]string
	WorkDir 	string
	FlagI 		bool
	FlagT		bool
}

func (dock Docker) start() []string {
	cmd := []string{"run"}

	if dock.FlagI {
		cmd = append(cmd, "-i")
	}

	if dock.FlagT {
		cmd = append(cmd, "-t")
	}

	if dock.WorkDir != "" {
		cmd = append(cmd, "-w", dock.WorkDir)
	}

	if dock.NetAlias != "" {
		cmd = append(cmd, "--net-alias", dock.NetAlias)
	}

	if dock.Net != "" {
		cmd = append(cmd, "--net", dock.Net)
	}

	if dock.Name != "" {
		cmd = append(cmd, "--name", dock.Name)
	}

	cmd = append(cmd, dock.Ports...)
	cmd = append(cmd, dock.Env)

	return append(cmd, dock.Image)
}
