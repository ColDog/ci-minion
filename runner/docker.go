package main

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
	FlagD		bool
}

func (dock Docker) start() []string {
	cmd := []string{"run"}

	if dock.FlagI {
		cmd = append(cmd, "-i")
	}

	if dock.FlagT {
		cmd = append(cmd, "-t")
	}

	if dock.FlagD {
		cmd = append(cmd, "-d")
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

	for _, env := range dock.Env {
		if env != "" {
			cmd = append(cmd, "-e", env)
		}
	}

	return append(cmd, dock.Image)
}
