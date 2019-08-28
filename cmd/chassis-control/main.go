package main

// chassis-control sends a chassis control command to a system, e.g. to power it
// on, or do a hard reset.

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gebn/bmc"
	"github.com/gebn/bmc/pkg/ipmi"

	"github.com/alecthomas/kingpin"
)

var (
	argBMCAddr = kingpin.Arg("addr", "IP[:port] of the BMC to control.").
			Required().
			String()
	argCommand = kingpin.Arg("command", "The command to send (on/off/cycle/reset/interrupt/softoff).").
			Required().
			String()
	flgUsername = kingpin.Flag("username", "The username to connect as.").
			Required().
			String()
	flgPassword = kingpin.Flag("password", "The password of the user to connect as.").
			Required().
			String()

	cmdControls = map[string]ipmi.ChassisControl{
		"off":       ipmi.ChassisControlPowerOff,
		"on":        ipmi.ChassisControlPowerOn,
		"cycle":     ipmi.ChassisControlPowerCycle,
		"reset":     ipmi.ChassisControlHardReset,
		"interrupt": ipmi.ChassisControlDiagnosticInterrupt,
		"softoff":   ipmi.ChassisControlSoftPowerOff,
	}
)

func lookupCommand(cmd string) (ipmi.ChassisControl, error) {
	if ctrl, ok := cmdControls[cmd]; ok {
		return ctrl, nil
	}
	return ipmi.ChassisControlPowerOff, fmt.Errorf("invalid command: %v", cmd)
}

func main() {
	kingpin.Parse()

	machine, err := bmc.DialV2(*argBMCAddr) // TODO change to Dial (need to implement v1.5 sessionless communication...)
	if err != nil {
		log.Fatal(err)
	}
	defer machine.Close()

	log.Printf("connected to %v over IPMI v%v", machine.Address(), machine.Version())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	sess, err := machine.NewSession(ctx, &bmc.SessionOpts{
		Username: *flgUsername,
		Password: []byte(*flgPassword),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close(ctx)

	cmd, err := lookupCommand(*argCommand)
	if err != nil {
		log.Fatal(err)
	}
	if err := sess.ChassisControl(ctx, cmd); err != nil {
		log.Fatal(err)
	}
}
