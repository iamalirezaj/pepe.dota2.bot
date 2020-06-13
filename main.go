package main

import (
	"context"
	"fmt"
	"github.com/MrJoshLab/go-dota2"
	"github.com/MrJoshLab/go-dota2/cso"
	"github.com/MrJoshLab/go-dota2/events"
	"github.com/MrJoshLab/go-dota2/protocol"
	"github.com/MrJoshLab/go-dota2/socache"
	"github.com/faceit/go-steam"
	"github.com/faceit/go-steam/netutil"
	"github.com/faceit/go-steam/protocol/steamlang"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"reflect"
	"time"
)

var (
	dota2Client *dota2.Dota2
	dota2GCconnected = false
)

func main() {

	log.SetFlags(log.Lshortfile)

	server := netutil.ParsePortAddr("162.254.196.83:27018")

	c := steam.NewClient()
	c.ConnectTo(server)

	log.Println("Connected to: ", server)

	handleSteamEvents(c)
}

func handleSteamEvents(c *steam.Client)  {

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	logger := logrus.NewEntry(logrus.New())
	dota2Client = dota2.New(c, logger)

	for event := range c.Events() {

		log.Println(reflect.TypeOf(event))

		switch event.(type) {
		case *events.ClientWelcomed:
			dota2GCconnected = true
		case *events.GCConnectionStatusChanged:
			gcConnStatus := event.(*events.GCConnectionStatusChanged)
			if gcConnStatus.NewState != protocol.GCConnectionStatus_GCConnectionStatus_HAVE_SESSION  {
				dota2GCconnected = false
			}
		case *steam.LoggedOffEvent:
			logOff := event.(*steam.LoggedOffEvent)
			log.Println(logOff)
		case *steam.FatalErrorEvent:
			log.Fatal(event)
		case *steam.ChatMsgEvent:
			msg := event.(*steam.ChatMsgEvent)
			log.Println(msg)
		case *steam.FriendStateEvent:
			fse := event.(*steam.FriendStateEvent)
			switch fse.Relationship {
			case steamlang.EFriendRelationship_None:
				log.Printf("Friend removed: [%s]", fse.SteamId.String())
			case steamlang.EFriendRelationship_RequestRecipient:
				log.Printf("New friend request: [%s]", fse.SteamId.String())
				if !fse.IsFriend() {
					c.Social.AddFriend(fse.SteamId)
					log.Printf("Accepted friend: [%s]", fse.SteamId.String())
				}
			}
		case *steam.ConnectedEvent:
			// Signing in
			c.Auth.LogOn(&steam.LogOnDetails{
				Username: os.Getenv("USERNAME"),
				Password: os.Getenv("PASSWORD"),
			})
		case error:
			fmt.Printf("Error: %v", event)
		case *steam.LoggedOnEvent:
			// Logged on
			c.Social.SetPersonaState(steamlang.EPersonaState_Online)
			c.GC.SetGamesPlayed(570)
			go connectToDota2GC()
		}
	}
}

func tryToConnectToDota2GC()  {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			if !dota2GCconnected {
				dota2Client.SayHello()
			}
		}
	}
}

func connectToDota2GC() {

	go tryToConnectToDota2GC()

	partyInviteCh, _, err := dota2Client.GetCache().SubscribeType(cso.PartyInvite)
	if err != nil {
		log.Println(err)
	}

	for {
		select {

		case e := <-partyInviteCh:

			log.Println(e)

			switch e.EventType {
			case socache.EventTypeCreate:

				log.Println(reflect.TypeOf(e.Object))

				party := e.Object.(*protocol.CSODOTAPartyInvite)
				dota2Client.RespondPartyInvite(party.GetGroupId(), true)
				dota2Client.SetMemberPartyCoach(true)

				resp, err := dota2Client.JoinChatChannel(
					context.Background(),
					"Party",
					protocol.DOTAChatChannelTypeT_DOTAChannelType_Party,
				)

				if err != nil {
					log.Println(err)
				}

				log.Println(resp)

				break
			}

		}
	}

}