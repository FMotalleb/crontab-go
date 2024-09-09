package connection

// reshapeEnviron modifies the environment variables for a given task.
// It allows overriding the global shell and shell arguments with task-specific values.
//
// Parameters:
//   - taskEnvironments: A task-specific environment variables.
//   - log: A logrus.Entry used for logging information about environment variable overrides.
//
// Returns:
//   - string: The shell to be used for the task, either the global shell or the overridden shell.
//   - []string: The shell arguments to be used for the task, either the global shell arguments or the overridden shell arguments.
//   - []string: The complete set of environment variables for the task, including any task-specific overrides.
// func reshapeEnviron(ctx CmdCtx, taskEnvironments map[string]string, log *logrus.Entry) (string, []string, []string) {
// 	shell := cmd.CFG.Shell
// 	shellArgs := strings.Split(cmd.CFG.ShellArgs[0], ":")
// 	env := os.Environ()
// 	log.Trace("Initial environment variables: ", env)
// 	for key, val := range taskEnvironments {
// 		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(key), val))
// 		oldValue := os.Getenv(key)
// 		log.Tracef("Adding environment variable: %s=%s, before this change was: `%s`", key, val, oldValue)
// 		switch strings.ToLower(key) {
// 		case "shell":
// 			log.Info("you've used `SHELL` env variable in command environments, overriding the global shell with:", val)
// 			shell = val
// 		case "shell_args":
// 			log.Info("you've used `SHELL_ARGS` env variable in command environments, overriding the global shell_args with: ", val)
// 			shellArgs = strings.Split(val, ":")
// 		case "shell_arg_compatibility":
// 			log.Info("you've used `SHELL_ARG_COMPATIBILITY` env variable in command environments, overriding the global shell_arg_compatibility with: ", val)
// 			env = append(env, fmt.Sprintf("%s=%s", "CRONTAB_GO_EVENT_ARGUMENTS", val))
// 			shellArgCompatibility = strings.Split(val, ":")
// 		}
// 	}
// 	log.Trace("Final environment variables: ", env)
// 	return shell, shellArgs, env
// }
