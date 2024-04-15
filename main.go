package main

import "fmt"

func main() {
	const githubUrl = "https://github.com/microsoft/vscode-remote-try-rust"
	devContainerCli := DevContainerCli{}
	devContainer, err := devContainerCli.Create(LaunchDevContainerRequest{githubUrl: githubUrl})

	if err != nil {
		fmt.Println(err)
		return
	}

	cmd := "cargo run"

	execOut, err := devContainerCli.Exec(ExecCmdRequest{devContainer: devContainer, cmd: cmd})

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(execOut)
}
