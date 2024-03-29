package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gr00by87/fst/config"
	"github.com/gr00by87/fst/vpn"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/xlzd/gotp"
	"golang.org/x/net/context"
	survey "gopkg.in/AlecAivazis/survey.v1"
	surveyCore "gopkg.in/AlecAivazis/survey.v1/core"
)

var (
	vpnDisconnect *bool

	// vpnCmd represents the vpn command.
	vpnCmd = &cobra.Command{
		Use:   "vpn",
		Short: "Connect to VPN using Pritunl client",
		Long:  "This subcommand allows to connect to VPN using the pin+otp authorization method. It requires Prituln application to be installed in the system.",
		Run:   runVPN,
	}
)

// init initializes the cobra command and flags.
func init() {
	rootCmd.AddCommand(vpnCmd)
	vpnDisconnect = vpnCmd.Flags().BoolP("disconnect", "d", false, "disconnect from VPN")
}

// runVPN executes the vpn command.
func runVPN(_ *cobra.Command, _ []string) {
	cfg, err := config.LoadFromFile()
	if err != nil {
		exitWithError(err)
	}

	pritunl, err := vpn.NewPritunl()
	if err != nil {
		exitWithError(err)
	}

	if *vpnDisconnect {
		disconnect(cfg, pritunl)
		return
	}

	connect(cfg, pritunl)
}

// connect connects to vpn.
func connect(cfg *config.Config, pritunl *vpn.Pritunl) {
	credentials := vpn.ConnectionCredentials{
		ID: cfg.VPNConfig.ProfileID,
	}

	connected, err := pritunl.IsConnected(credentials.ID)
	if err != nil {
		exitWithError(err)
	}
	if connected {
		fmt.Println(info, "Connection to VPN already established")
		return
	}

	surveyCore.QuestionIcon = "🔒"
	prompts := []*survey.Question{
		{
			Name:   "Pin",
			Prompt: &survey.Password{Message: "Enter Pin:"},
			Validate: func(val interface{}) error {
				if len(val.(string)) < 1 {
					return errors.New("Pin cannot be empty")
				}
				return nil
			},
		},
	}

	if cfg.VPNConfig.OTPSecret == "" {
		prompts = append(prompts, &survey.Question{
			Name:     "OTP",
			Prompt:   &survey.Input{Message: "Enter OTP Code:"},
			Validate: validateLength("OTP Code", 6),
		})
	} else {
		credentials.OTP = gotp.NewDefaultTOTP(cfg.VPNConfig.OTPSecret).Now()
	}

	if err := survey.Ask(prompts, &credentials); err != nil {
		exitWithError(err)
	}

	events := make(chan vpn.Event, 100)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	unsubscribe, err := pritunl.GetEvents(ctx, events)
	if err != nil {
		exitWithError(err)
	}
	defer unsubscribe()

	err = pritunl.Connect(credentials)
	if err != nil {
		exitWithError(err)
	}
	fmt.Println(info, "Connecting to VPN...")

	for event := range events {
		if event.Err != nil {
			exitWithError(event.Err)
		}

		msgType := gjson.GetBytes(event.Message, "type").String()
		switch {
		case strings.HasSuffix(msgType, "_error"):
			fmt.Println(failure, "Connection failed:", strings.ReplaceAll(msgType, "_", " "))
			return
		case msgType == "connected":
			fmt.Println(success, "Connected")
			return
		}
	}
}

// disconnect disconnects from vpn.
func disconnect(cfg *config.Config, pritunl *vpn.Pritunl) {
	connected, err := pritunl.IsConnected(cfg.VPNConfig.ProfileID)
	if err != nil {
		exitWithError(err)
	}
	if !connected {
		fmt.Println(info, "Connection to VPN not established")
		return
	}

	events := make(chan vpn.Event, 100)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	unsubscribe, err := pritunl.GetEvents(ctx, events)
	if err != nil {
		exitWithError(err)
	}
	defer unsubscribe()

	if err := pritunl.Disconnect(cfg.VPNConfig.ProfileID); err != nil {
		exitWithError(err)
	}
	fmt.Println(info, "Disconnecting from VPN...")

	for event := range events {
		if event.Err != nil {
			exitWithError(event.Err)
		}

		msgType := gjson.GetBytes(event.Message, "type").String()
		if msgType == "disconnected" {
			fmt.Println(success, "Disconnected")
			return
		}
	}
}
