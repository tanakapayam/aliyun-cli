/*
 * Copyright (C) 2017-2018 Alibaba Group Holding Limited
 */
package config

import (
	"fmt"
	"github.com/aliyun/aliyun-cli/cli"
	"github.com/aliyun/aliyun-cli/i18n"
	"io"
	"io/ioutil"
	"strings"
)

func NewConfigureCommand() *cli.Command {
	c := &cli.Command{
		Name: "configure",
		Short: i18n.T(
			"configure credential and settings",
			"配置身份认证和其他信息"),
		Usage: "configure --mode <AuthenticateMode> --profile <profileName>",
		Run: func(ctx *cli.Context, args []string) error {
			if len(args) > 0 {
				return cli.NewInvalidCommandError(args[0], ctx)
			}
			profileName, _ := ProfileFlag.GetValue()
			mode, _ := ModeFlag.GetValue()

			return doConfigure(ctx.Writer(), profileName, mode)
		},
	}

	c.Flags().Add(ProfileFlag)
	c.Flags().Add(ModeFlag)

	c.AddSubCommand(NewConfigureGetCommand(w))
	c.AddSubCommand(NewConfigureSetCommand(w))
	c.AddSubCommand(NewConfigureListCommand(w))
	c.AddSubCommand(NewConfigureDeleteCommand(w))
	return c
}

func doConfigure(w io.Writer, profileName string, mode string) error {
	conf, err := LoadConfiguration(w)
	if err != nil {
		return err
	}

	cp, ok := conf.GetProfile(profileName)
	if !ok {
		cp = conf.NewProfile(profileName)
	}

	cli.Printf(w, "Configuring profile '%s' in '%s' authenticate mode...\n", profileName, mode)

	if mode != "" {
		switch AuthenticateMode(mode) {
		case AK:
			cp.Mode = AK
			configureAK(w, &cp)
		case StsToken:
			cp.Mode = StsToken
			configureStsToken(w, &cp)
		case RamRoleArn:
			cp.Mode = RamRoleArn
			configureRamRoleArn(w, &cp)
		case EcsRamRole:
			cp.Mode = EcsRamRole
			configureEcsRamRole(w, &cp)
		case RsaKeyPair:
			cp.Mode = RsaKeyPair
			configureRsaKeyPair(w, &cp)
		default:
			return fmt.Errorf("unexcepted authenticate mode: %s", mode)
		}
	} else {
		configureAK(w, &cp)
	}

	//
	// configure common
	cli.Printf(w, "Default Region Id [%s]: ", cp.RegionId)
	cp.RegionId = ReadInput(cp.RegionId)
	cli.Printf(w, "Default Output Format [%s]: json (Only support json))\n", cp.OutputFormat)

	// cp.OutputFormat = ReadInput(cp.OutputFormat)
	cp.OutputFormat = "json"

	cli.Printf(w, "Default Language [zh|en] %s: ", cp.Language)

	cp.Language = ReadInput(cp.Language)
	if cp.Language != "zh" && cp.Language != "en" {
		cp.Language = "en"
	}

	//fmt.Printf("User site: [china|international|japan] %s", cp.Site)
	//cp.Site = ReadInput(cp.Site)

	cli.Printf(w, "Saving profile[%s] ...", profileName)

	conf.PutProfile(cp)
	conf.CurrentProfile = cp.Name
	err = SaveConfiguration(conf)

	if err != nil {
		return err
	}
	cli.Printf(w, "Done.\n")

	DoHello(w, &cp)
	return nil
}

func configureAK(w io.Writer, cp *Profile) error {
	cli.Printf(w, "Access Key Id [%s]: ", MosaicString(cp.AccessKeyId, 3))
	cp.AccessKeyId = ReadInput(cp.AccessKeyId)
	cli.Printf(w, "Access Key Secret [%s]: ", MosaicString(cp.AccessKeySecret, 3))
	cp.AccessKeySecret = ReadInput(cp.AccessKeySecret)
	return nil
}

func configureStsToken(w io.Writer, cp *Profile) error {
	err := configureAK(w, cp)
	if err != nil {
		return err
	}
	cli.Printf(w, "Sts Token [%s]: ", cp.StsToken)
	cp.StsToken = ReadInput(cp.StsToken)
	return nil
}

func configureRamRoleArn(w io.Writer, cp *Profile) error {
	err := configureAK(w, cp)
	if err != nil {
		return err
	}
	cli.Printf(w, "Ram Role Arn [%s]: ", cp.RamRoleArn)
	cp.RamRoleArn = ReadInput(cp.RamRoleArn)
	cli.Printf(w, "Role Session Name [%s]: ", cp.RoleSessionName)
	cp.RoleSessionName = ReadInput(cp.RoleSessionName)
	cp.ExpiredSeconds = 900
	return nil
}

func configureEcsRamRole(w io.Writer, cp *Profile) error {
	cli.Printf(w, "Ecs Ram Role [%s]: ", cp.RamRoleName)
	cp.RamRoleName = ReadInput(cp.RamRoleName)
	return nil
}

func configureRsaKeyPair(w io.Writer, cp *Profile) error {
	cli.Printf(w, "Rsa Private Key File: ")
	keyFile := ReadInput("")
	buf, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("read key file %s failed %v", keyFile, err)
	}
	cp.PrivateKey = string(buf)
	cli.Printf(w, "Rsa Key Pair Name: ")
	cp.KeyPairName = ReadInput("")
	cp.ExpiredSeconds = 900
	return nil
}

func ReadInput(defaultValue string) string {
	var s string
	fmt.Scanf("%s\n", &s)
	if s == "" {
		return defaultValue
	}
	return s
}

func MosaicString(s string, lastChars int) string {
	r := len(s) - lastChars
	if r > 0 {
		return strings.Repeat("*", r) + s[r:]
	} else {
		return strings.Repeat("*", len(s))
	}
}

func GetLastChars(s string, lastChars int) string {
	r := len(s) - lastChars
	if r > 0 {
		return s[r:]
	} else {
		return strings.Repeat("*", len(s))
	}
}
