package cmd

import (
	"fmt"
	"github.com/nwtgck/handy-sshd"
	"github.com/nwtgck/handy-sshd/version"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/exp/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"bufio"
)

type flagType struct {
	//dnsServer    string
	showsVersion  bool
	sshHost       string
	sshPort       uint16
	sshUnixSocket string
	sshShell      string
	sshUsers      []string
	sshAuthKeys   []string

	allowTcpipForward       bool
	allowDirectTcpip        bool
	allowExecute            bool
	allowSftp               bool
	allowStreamlocalForward bool
	allowDirectStreamlocal  bool
}

type permissionFlagType = struct {
	name    string
	flagPtr *bool
}

type sshUser struct {
	name     string
	password string
}

func init() {
	cobra.OnInitialize()
}

func RootCmd() *cobra.Command {
	var flag flagType
	allPermissionFlags := []permissionFlagType{
		{name: "tcpip-forward", flagPtr: &flag.allowTcpipForward},
		{name: "direct-tcpip", flagPtr: &flag.allowDirectTcpip},
		{name: "execute", flagPtr: &flag.allowExecute},
		{name: "sftp", flagPtr: &flag.allowSftp},
		{name: "streamlocal-forward", flagPtr: &flag.allowStreamlocalForward},
		{name: "direct-streamlocal", flagPtr: &flag.allowDirectStreamlocal},
	}
	rootCmd := cobra.Command{
		Use:          os.Args[0],
		Short:        "cpolar_friend_1_ssl",
		Long:         "A SSL-like tool implement",
		SilenceUsage: true,
		Example: `Sorry, but you should refer to original project. This safe implement doesn't contain example usages to reduce the attack surface.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootRunEWithExtra(cmd, args, &flag, allPermissionFlags)
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&flag.showsVersion, "version", "v", false, "show version")
	rootCmd.PersistentFlags().StringVarP(&flag.sshHost, "host", "", "", "Listen (e.g. 127.0.0.1)")
	rootCmd.PersistentFlags().Uint16VarP(&flag.sshPort, "port", "p", 2222, "port")
	// NOTE: long name 'unix-socket' is from curl (ref: https://curl.se/docs/manpage.html)
	rootCmd.PersistentFlags().StringVarP(&flag.sshUnixSocket, "unix-socket", "", "", "Unix socket")
	rootCmd.PersistentFlags().StringVarP(&flag.sshShell, "shell", "", "", "Shell")
	//rootCmd.PersistentFlags().StringVar(&flag.dnsServer, "dns-server", "", "DNS server (e.g. 1.1.1.1:53)")
	rootCmd.PersistentFlags().StringArrayVarP(&flag.sshUsers, "user", "u", nil, `user and pass seperated by ":"`)
	rootCmd.PersistentFlags().StringArrayVarP(&flag.sshAuthKeys, "keys", "k", nil, "A-Key file")

	// Permission flags
	rootCmd.PersistentFlags().BoolVarP(&flag.allowTcpipForward, "allow-tcpip-forward", "", false, "Allow \"-R\" technique")
	rootCmd.PersistentFlags().BoolVarP(&flag.allowDirectTcpip, "allow-direct-tcpip", "", false, "Allow \"-L\" and \"-D\" technique")
	rootCmd.PersistentFlags().BoolVarP(&flag.allowExecute, "allow-execute", "", false, "Can login")
	rootCmd.PersistentFlags().BoolVarP(&flag.allowSftp, "allow-sftp", "", false, "File management")
	rootCmd.PersistentFlags().BoolVarP(&flag.allowStreamlocalForward, "allow-streamlocal-forward", "", false, "\"-R\" technique permission but for unix sockets")
	rootCmd.PersistentFlags().BoolVarP(&flag.allowDirectStreamlocal, "allow-direct-streamlocal", "", false, "Direct unix socket conn \"-L\"")

	return &rootCmd
}

func rootRunEWithExtra(cmd *cobra.Command, args []string, flag *flagType, allPermissionFlags []permissionFlagType) error {
	if flag.showsVersion {
		fmt.Fprintln(cmd.OutOrStdout(), version.Version)
		return nil
	}
	logger := slog.Default()

	// Allow all permissions if all permission is not set
	{
		allPermissionFalse := true
		for _, permissionFlag := range allPermissionFlags {
			allPermissionFalse = allPermissionFalse && !*permissionFlag.flagPtr
		}
		if allPermissionFalse {
			for _, permissionFlag := range allPermissionFlags {
				*permissionFlag.flagPtr = true
			}
		}
	}

	sshServer := &handy_sshd.Server{
		Logger:                  logger,
		AllowTcpipForward:       flag.allowTcpipForward,
		AllowDirectTcpip:        flag.allowDirectTcpip,
		AllowExecute:            flag.allowExecute,
		AllowSftp:               flag.allowSftp,
		AllowStreamlocalForward: flag.allowStreamlocalForward,
		AllowDirectStreamlocal:  flag.allowDirectStreamlocal,
	}
	var sshUsers []sshUser
	for _, u := range flag.sshUsers {
		splits := strings.SplitN(u, ":", 2)
		if len(splits) != 2 {
			return fmt.Errorf("invalid user format: %s", u)
		}
		sshUsers = append(sshUsers, sshUser{name: splits[0], password: splits[1]})
	}
	authorizedKeys := make(map[string]bool)
	for _, k := range flag.sshAuthKeys {
		hand, err := os.Open(k)
		if err != nil { continue }
		scanner := bufio.NewScanner(hand)
		for scanner.Scan() {
			key := scanner.Text()
			pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
			if err != nil { continue }
			authorizedKeys[string(ssh.FingerprintSHA256(pub))] = true
		}
	}
	// (base: https://gist.github.com/jpillora/b480fde82bff51a06238)
	sshConfig := &ssh.ServerConfig{
		PublicKeyCallback: func(metadata ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
            if authorizedKeys[string(ssh.FingerprintSHA256(key))] {
                return nil, nil
            }
            return nil, fmt.Errorf("%q P-Key failed", metadata.User())
        },
		//Define a function to run when a client attempts a password login
		PasswordCallback: func(metadata ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			for _, user := range sshUsers {
				// No auth required
				if user.name == metadata.User() && user.password == string(pass) {
					return nil, nil
				}
			}
			return nil, fmt.Errorf("%q pass failed", metadata.User())
		},
		NoClientAuth: true,
		NoClientAuthCallback: func(metadata ssh.ConnMetadata) (*ssh.Permissions, error) {
			for _, user := range sshUsers {
				// No auth required
				if user.name == metadata.User() && user.password == "" {
					return nil, nil
				}
			}
			return nil, fmt.Errorf("NO_AUTH %s", metadata.User())
		},
	}
	// TODO: specify priv_key by flags
	pri, err := ssh.ParsePrivateKey([]byte(defaultHostKeyPem))
	if err != nil {
		return err
	}
	sshConfig.AddHostKey(pri)

	var ln net.Listener
	if flag.sshUnixSocket == "" {
		address := net.JoinHostPort(flag.sshHost, strconv.Itoa(int(flag.sshPort)))
		ln, err = net.Listen("tcp", address)
		if err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("Running on %s...", address))
	} else {
		ln, err = net.Listen("unix", flag.sshUnixSocket)
		if err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("Running on %s...", flag.sshUnixSocket))
	}
	defer ln.Close()

	showPermissions(logger, allPermissionFlags)

	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Error("TCP failed", "err", err)
			continue
		}
		sshConn, chans, reqs, err := ssh.NewServerConn(conn, sshConfig)
		if err != nil {
			logger.Info("Handshake failed", "err", err)
			conn.Close()
			continue
		}
		logger.Info("new conn", "remote_address", sshConn.RemoteAddr(), "client_version", "<hidden>")
		go sshServer.HandleGlobalRequests(sshConn, reqs)
		go sshServer.HandleChannels(flag.sshShell, chans)
	}
}

func showPermissions(logger *slog.Logger, allPermissionFlags []permissionFlagType) {
	var allowedList []string
	var notAllowedList []string
	for _, permissionFlag := range allPermissionFlags {
		if *permissionFlag.flagPtr {
			allowedList = append(allowedList, `"`+permissionFlag.name+`"`)
		} else {
			notAllowedList = append(notAllowedList, `"`+permissionFlag.name+`"`)
		}
	}
	showList := func(l []string) string {
		if len(l) == 0 {
			return "none"
		}
		return strings.Join(l, ", ")
	}
	logger.Info(fmt.Sprintf("allowed: %s", showList(allowedList)))
	logger.Info(fmt.Sprintf("NOT allowed: %s", showList(notAllowedList)))
}
