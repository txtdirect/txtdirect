// Package txtdirectmain contains the functions for starting TXTDirect.
package txtdirectmain

// func init() {
// 	caddy.Quiet = true // don't show init stuff from caddy
// 	setVersion()

// 	caddy.RegisterCaddyfileLoader("flag", caddy.LoaderFunc(confLoader))
// 	caddy.SetDefaultCaddyfileLoader("default", caddy.LoaderFunc(defaultLoader))

// 	caddy.AppName = TXTDirectName
// 	caddy.AppVersion = TXTDirectVersion
// }

// // Run is TXTDirect's main() function.
// func Run() {
// 	flag.Parse()
// 	caddy.TrapSignals()

// 	if err := getFlags(); err != nil {
// 		log.Fatalf("[txtdirect]: Couldn't parse the flags: %s", err.Error())
// 	}

// 	if version {
// 		showVersion()
// 		os.Exit(0)
// 	}

// 	// Get TXTDirect config input
// 	caddyfile, err := caddy.LoadCaddyfile("http")
// 	if err != nil {
// 		mustLogFatal(err)
// 	}

// 	// Start your engines
// 	instance, err := caddy.Start(caddyfile)
// 	if err != nil {
// 		mustLogFatal(err)
// 	}

// 	// Twiddle your thumbs
// 	instance.Wait()
// }

// // mustLogFatal wraps log.Fatal() in a way that ensures the
// // output is always printed to stderr so the user can see it
// // if the user is still there, even if the process log was not
// // enabled. If this process is an upgrade, however, and the user
// // might not be there anymore, this just logs to the process
// // log and exits.
// func mustLogFatal(args ...interface{}) {
// 	if !caddy.IsUpgrade() {
// 		log.SetOutput(os.Stderr)
// 	}
// 	log.Fatal(args...)
// }

// // confLoader loads the Caddyfile using the -conf flag.
// func confLoader(serverType string) (caddy.Input, error) {
// 	if conf == "" {
// 		return nil, nil
// 	}

// 	if conf == "stdin" {
// 		return caddy.CaddyfileFromPipe(os.Stdin, serverType)
// 	}

// 	contents, err := ioutil.ReadFile(conf)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return caddy.CaddyfileInput{
// 		Contents:       contents,
// 		Filepath:       conf,
// 		ServerTypeName: serverType,
// 	}, nil
// }

// // defaultLoader loads the TXTDirect config from the current working directory.
// func defaultLoader(serverType string) (caddy.Input, error) {
// 	contents, err := ioutil.ReadFile(caddy.DefaultConfigFile)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return caddy.CaddyfileInput{
// 		Contents:       contents,
// 		Filepath:       caddy.DefaultConfigFile,
// 		ServerTypeName: serverType,
// 	}, nil
// }

// // showVersion prints the version that is starting.
// func showVersion() {
// 	fmt.Print(versionString())
// 	fmt.Print(releaseString())
// 	if devBuild && gitShortStat != "" {
// 		fmt.Printf("%s\n%s\n", gitShortStat, gitFilesModified)
// 	}
// }

// // versionString returns the TXTDirect version as a string.
// func versionString() string {
// 	return fmt.Sprintf("%s-%s\n", caddy.AppName, caddy.AppVersion)
// }

// // releaseString returns the release information related to TXTDirect version:
// // <OS>/<ARCH>, <go version>, <commit>
// // e.g.,
// // linux/amd64, go1.8.3, a6d2d7b5
// func releaseString() string {
// 	return fmt.Sprintf("%s/%s, %s, %s\n", runtime.GOOS, runtime.GOARCH, runtime.Version(), GitCommit)
// }

// // setVersion figures out the version information
// // based on variables set by -ldflags.
// func setVersion() {
// 	// A development build is one that's not at a tag or has uncommitted changes
// 	devBuild = gitTag == "" || gitShortStat != ""

// 	// Only set the appVersion if -ldflags was used
// 	if gitNearestTag != "" || gitTag != "" {
// 		if devBuild && gitNearestTag != "" {
// 			appVersion = fmt.Sprintf("%s (+%s %s)", strings.TrimPrefix(gitNearestTag, "v"), GitCommit, buildDate)
// 		} else if gitTag != "" {
// 			appVersion = strings.TrimPrefix(gitTag, "v")
// 		}
// 	}
// }

// // Flags that control program flow or startup
// var (
// 	conf    string
// 	version bool
// )

// // Build information obtained with the help of -ldflags
// var (
// 	appVersion = "(untracked dev build)" // inferred at startup
// 	devBuild   = true                    // inferred at startup

// 	buildDate        string // date -u
// 	gitTag           string // git describe --exact-match HEAD 2> /dev/null
// 	gitNearestTag    string // git describe --abbrev=0 --tags HEAD
// 	gitShortStat     string // git diff-index --shortstat
// 	gitFilesModified string // git diff-index --name-only HEAD

// 	// Gitcommit contains the commit where we built TXTDirect from.
// 	GitCommit string
// )

// func getFlags() error {
// 	versionFlag := flag.Lookup("version")
// 	if versionFlag != nil {
// 		versionVal, err := strconv.ParseBool(versionFlag.Value.String())
// 		if err != nil {
// 			return fmt.Errorf("Coudln't parse the -version flag: %s", err.Error())
// 		}
// 		version = versionVal
// 	}

// 	return nil
// }
