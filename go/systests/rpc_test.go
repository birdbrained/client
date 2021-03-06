// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package systests

// Test various RPCs that are used mainly in other clients but not by the CLI.

import (
	"regexp"
	"strings"
	"testing"

	"github.com/keybase/client/go/client"
	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol/keybase1"
	"github.com/keybase/client/go/service"
	"github.com/keybase/client/go/teams"
	"github.com/stretchr/testify/require"
	context "golang.org/x/net/context"
)

func TestRPCs(t *testing.T) {
	tc := setupTest(t, "rpcs")
	tc2 := cloneContext(tc)

	libkb.G.LocalDb = nil

	defer tc.Cleanup()

	stopCh := make(chan error)
	svc := service.NewService(tc.G, false)
	startCh := svc.GetStartChannel()
	go func() {
		err := svc.Run()
		if err != nil {
			t.Logf("Running the service produced an error: %v", err)
		}
		stopCh <- err
	}()

	<-startCh

	// Add test RPC methods here.
	testIdentifyResolve3(t, tc2.G)
	testCheckInvitationCode(t, tc2.G)
	testLoadAllPublicKeysUnverified(t, tc2.G)
	testLoadUserWithNoKeys(t, tc2.G)
	testCheckDevicesForUser(t, tc2.G)
	testIdentify2(t, tc2.G)
	testMerkle(t, tc2.G)
	testIdentifyLite(t)

	if err := client.CtlServiceStop(tc2.G); err != nil {
		t.Fatal(err)
	}

	// If the server failed, it's also an error
	if err := <-stopCh; err != nil {
		t.Fatal(err)
	}
}

func testIdentifyResolve3(t *testing.T, g *libkb.GlobalContext) {

	cli, err := client.GetIdentifyClient(g)
	if err != nil {
		t.Fatalf("failed to get new identifyclient: %v", err)
	}

	// We don't want to hit the cache, since the previous lookup never hit the
	// server.  For Resolve3, we have to, since we need a username.  So test that
	// here.
	if res, err := cli.Resolve3(context.TODO(), "uid:eb72f49f2dde6429e5d78003dae0c919"); err != nil {
		t.Fatalf("Resolve failed: %v\n", err)
	} else if res.Name != "t_tracy" {
		t.Fatalf("Wrong username: %s != 't_tracy", res.Name)
	}

	if res, err := cli.Resolve3(context.TODO(), "t_tracy@rooter"); err != nil {
		t.Fatalf("Resolve3 failed: %v\n", err)
	} else if res.Name != "t_tracy" {
		t.Fatalf("Wrong name: %s != 't_tracy", res.Name)
	} else if !res.Id.AsUserOrBust().Equal(keybase1.UID("eb72f49f2dde6429e5d78003dae0c919")) {
		t.Fatalf("Wrong uid for tracy: %s\n", res.Id)
	}

	if _, err := cli.Resolve3(context.TODO(), "foobag@rooter"); err == nil {
		t.Fatalf("expected an error on a bad resolve, but got none")
	} else if _, ok := err.(libkb.ResolutionError); !ok {
		t.Fatalf("Wrong error: wanted type %T but got (%v, %T)", libkb.ResolutionError{}, err, err)
	}

	if res, err := cli.Resolve3(context.TODO(), "t_tracy"); err != nil {
		t.Fatalf("Resolve3 failed: %v\n", err)
	} else if res.Name != "t_tracy" {
		t.Fatalf("Wrong name: %s != 't_tracy", res.Name)
	} else if !res.Id.AsUserOrBust().Equal(keybase1.UID("eb72f49f2dde6429e5d78003dae0c919")) {
		t.Fatalf("Wrong uid for tracy: %s\n", res.Id)
	}
}

func testCheckInvitationCode(t *testing.T, g *libkb.GlobalContext) {
	cli, err := client.GetSignupClient(g)
	if err != nil {
		t.Fatalf("failed to get a signup client: %v", err)
	}

	err = cli.CheckInvitationCode(context.TODO(), keybase1.CheckInvitationCodeArg{InvitationCode: libkb.TestInvitationCode})
	if err != nil {
		t.Fatalf("Did not expect an error code, but got: %v", err)
	}
	err = cli.CheckInvitationCode(context.TODO(), keybase1.CheckInvitationCodeArg{InvitationCode: "eeoeoeoe333o3"})
	if _, ok := err.(libkb.BadInvitationCodeError); !ok {
		t.Fatalf("Expected an error code, but got %T %v", err, err)
	}
}

func testLoadAllPublicKeysUnverified(t *testing.T, g *libkb.GlobalContext) {

	cli, err := client.GetUserClient(g)
	if err != nil {
		t.Fatalf("failed to get user client: %s", err)
	}

	// t_rosetta
	arg := keybase1.LoadAllPublicKeysUnverifiedArg{Uid: keybase1.UID("b8939251cb3d367e68587acb33a64d19")}
	res, err := cli.LoadAllPublicKeysUnverified(context.TODO(), arg)
	if err != nil {
		t.Fatalf("failed to make load keys call: %s", err)
	}

	if len(res) != 3 {
		t.Fatalf("wrong amount of keys loaded: %d != %d", len(res), 3)
	}

	keys := map[keybase1.KID]bool{
		keybase1.KID("0101fe1183765f256289427d6943cd8bab3b5fe095bcdd27f031ed298da523efd3120a"): true,
		keybase1.KID("0101b5839c4ccaa9d03b3016b9aa73a7e3eafb067f9c86c07a6f2f79cb8558b1c97f0a"): true,
		keybase1.KID("0101188ee7e63ccbd05af498772ab2975ee29df773240d17dde09aecf6c132a5a9a60a"): true,
	}

	for _, key := range res {
		if _, ok := keys[key.KID]; !ok {
			t.Fatalf("unknown key in response: %s", key.KID)
		}
	}
}

func testLoadUserWithNoKeys(t *testing.T, g *libkb.GlobalContext) {
	// The LoadUser class in libkb returns an error by default if the user in
	// question has no keys. The RPC methods that wrap it should suppress this
	// error, by setting the PublicKeyOptional flag.

	cli, err := client.GetUserClient(g)
	if err != nil {
		t.Fatalf("failed to get a user client: %v", err)
	}

	// Check the LoadUserByName RPC. t_ellen is a test user with no keys.
	loadUserByNameArg := keybase1.LoadUserByNameArg{Username: "t_ellen"}
	tEllen, err := cli.LoadUserByName(context.TODO(), loadUserByNameArg)
	if err != nil {
		t.Fatal(err)
	}
	if tEllen.Username != "t_ellen" {
		t.Fatalf("expected t_ellen, saw %s", tEllen.Username)
	}

	// Check the LoadUser RPC.
	loadUserArg := keybase1.LoadUserArg{Uid: tEllen.Uid}
	tEllen2, err := cli.LoadUser(context.TODO(), loadUserArg)
	if err != nil {
		t.Fatal(err)
	}
	if tEllen2.Username != "t_ellen" {
		t.Fatalf("expected t_ellen, saw %s", tEllen2.Username)
	}
}

func testCheckDevicesForUser(t *testing.T, g *libkb.GlobalContext) {
	cli, err := client.GetDeviceClient(g)
	if err != nil {
		t.Fatalf("failed to get a device client: %v", err)
	}
	err = cli.CheckDeviceNameForUser(context.TODO(), keybase1.CheckDeviceNameForUserArg{
		Username:   "t_frank",
		Devicename: "bad $ device $ name",
	})
	if _, ok := err.(libkb.DeviceBadNameError); !ok {
		t.Fatalf("wanted a bad device name error; got %v", err)
	}
	err = cli.CheckDeviceNameForUser(context.TODO(), keybase1.CheckDeviceNameForUserArg{
		Username:   "t_frank",
		Devicename: "go c lient",
	})
	if _, ok := err.(libkb.DeviceNameInUseError); !ok {
		t.Fatalf("wanted a name in use error; got %v", err)
	}
}

func testIdentify2(t *testing.T, g *libkb.GlobalContext) {

	cli, err := client.GetIdentifyClient(g)
	if err != nil {
		t.Fatalf("failed to get new identifyclient: %v", err)
	}

	_, err = cli.Identify2(context.TODO(), keybase1.Identify2Arg{
		UserAssertion:    "t_alice",
		IdentifyBehavior: keybase1.TLFIdentifyBehavior_CHAT_GUI,
	})
	if err != nil {
		t.Fatalf("Identify2 failed: %v\n", err)
	}

	_, err = cli.Identify2(context.TODO(), keybase1.Identify2Arg{
		UserAssertion:    "t_weriojweroi",
		IdentifyBehavior: keybase1.TLFIdentifyBehavior_CHAT_GUI,
	})
	if _, ok := err.(libkb.NotFoundError); !ok {
		t.Fatalf("Expected a not-found error, but got: %v (%T)", err, err)
	}
}

func testMerkle(t *testing.T, g *libkb.GlobalContext) {

	cli, err := client.GetMerkleClient(g)
	if err != nil {
		t.Fatalf("failed to get new merkle client: %v", err)
	}

	root, err := cli.GetCurrentMerkleRoot(context.TODO(), int(-1))
	if err != nil {
		t.Fatalf("GetCurrentMerkleRoot failed: %v\n", err)
	}
	if root.Root.Seqno <= keybase1.Seqno(0) {
		t.Fatalf("Failed basic sanity check")
	}
}

func testIdentifyLite(t *testing.T) {

	tt := newTeamTester(t)
	defer tt.cleanup()

	tt.addUser("abc")
	teamName := tt.users[0].createTeam()
	g := tt.users[0].tc.G

	t.Logf("make a team")
	team, err := GetTeamForTestByStringName(context.Background(), g, teamName)
	require.NoError(t, err)

	getTeamName := func(teamID keybase1.TeamID) keybase1.TeamName {
		team, err := teams.Load(context.Background(), g, keybase1.LoadTeamArg{
			ID: teamID,
		})
		require.NoError(t, err)
		return team.Name()
	}

	t.Logf("make an implicit team")
	iTeamCreateName := strings.Join([]string{tt.users[0].username, "bob@github"}, ",")
	iTeamID, _, err := teams.LookupOrCreateImplicitTeam(context.TODO(), g, iTeamCreateName, false /*isPublic*/)
	require.NoError(t, err)
	iTeamImpName := getTeamName(iTeamID)
	require.True(t, iTeamImpName.IsImplicit())
	require.NoError(t, err)

	cli, err := client.GetIdentifyClient(g)
	require.NoError(t, err, "failed to get new identifyclient")

	// test ok assertions
	var units = []struct {
		assertion string
		resID     keybase1.TeamID
		resName   string
	}{
		{
			assertion: "t_alice",
			resName:   "t_alice",
		}, {
			assertion: "team:" + teamName,
			resID:     team.ID,
			resName:   teamName,
		}, {
			assertion: "tid:" + team.ID.String(),
			resID:     team.ID,
			resName:   teamName,
		},
	}
	for _, unit := range units {
		res, err := cli.IdentifyLite(context.Background(), keybase1.IdentifyLiteArg{Assertion: unit.assertion})
		require.NoError(t, err, "IdentifyLite (%s) failed", unit.assertion)

		if len(unit.resID) > 0 {
			require.Equal(t, unit.resID.String(), res.Ul.Id.String())
		}

		if len(unit.resName) > 0 {
			require.Equal(t, unit.resName, res.Ul.Name)
		}
	}

	// test identify by assertion and id
	assertions := []string{"team:" + teamName, "tid:" + team.ID.String()}
	for _, assertion := range assertions {
		_, err := cli.IdentifyLite(context.Background(), keybase1.IdentifyLiteArg{Id: team.ID.AsUserOrTeam(), Assertion: assertion})
		require.NoError(t, err, "IdentifyLite by assertion and id (%s)", assertion)
	}

	// test identify by id only
	_, err = cli.IdentifyLite(context.Background(), keybase1.IdentifyLiteArg{Id: team.ID.AsUserOrTeam()})
	require.NoError(t, err, "IdentifyLite id only")

	// test invalid user format
	_, err = cli.IdentifyLite(context.Background(), keybase1.IdentifyLiteArg{Assertion: "__t_alice"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "bad keybase username")

	// test team read error
	assertions = []string{"team:jwkj22111z"}
	for _, assertion := range assertions {
		_, err := cli.IdentifyLite(context.Background(), keybase1.IdentifyLiteArg{Assertion: assertion})
		aerr, ok := err.(libkb.AppStatusError)
		if ok {
			if aerr.Code != libkb.SCTeamNotFound {
				t.Fatalf("app status code: %d, expected %d", aerr.Code, libkb.SCTeamNotFound)
			}
		} else {
			require.True(t, regexp.MustCompile("Team .* does not exist").MatchString(err.Error()),
				"Expected an AppStatusError or team-does-not-exist for %s, but got: %v (%T)", assertion, err, err)
		}
	}

	// test not found assertions
	assertions = []string{"t_weriojweroi"}
	for _, assertion := range assertions {
		_, err := cli.IdentifyLite(context.Background(), keybase1.IdentifyLiteArg{Assertion: assertion})
		if _, ok := err.(libkb.NotFoundError); !ok {
			t.Fatalf("assertion %s, error: %s (%T), expected libkb.NotFoundError", assertion, err, err)
		}
	}
}
