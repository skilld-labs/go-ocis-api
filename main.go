package main

import (
	"fmt"
	"reflect"
	"encoding/xml"

	"main/docs"
	"github.com/davecgh/go-spew/spew"
	libregraph "github.com/owncloud/libre-graph-api-go"
	webdav "github.com/studio-b12/gowebdav"
)

const (
	// UnifiedRoleViewerID Unified role viewer id.
	UnifiedRoleViewerID = "b1e2218d-eef8-4d4c-b82d-0f1a1b48f3b5"
)

const (
	UsersMainSpace = "User Homes"
	UserName = "test"
)

func main() {

	cfg := docs.Config{
		Host: "***",
		Username: "***",
		Password: "***",
	}

	client := docs.New(cfg)

	// Test GraphAPI: get current user.
// 	response, _, err := client.GraphAPI.MeUserApi.GetOwnUser(client.GraphAPICtx).Execute()
// 	spew.Dump(response, err)

	// Test GraphAPI: get all drives where the current user is a regular member of.
// 	filter := "driveType eq 'project'"
// 	response, _, err := client.GraphAPI.MeDrivesApi.ListMyDrives(client.GraphAPICtx).Filter(filter).Execute()
// 	spew.Dump(response, err)

	// Test GraphAPI: Create a new drive of a specific type.
// 	drive := *libregraph.NewDrive(UsersMainSpace)
// 	drive.DriveType = stringToPointer("project")
// 	spew.Dump(drive)
// 	response, _, err := client.GraphAPI.DrivesApi.CreateDrive(client.GraphAPICtx).Drive(drive).Execute()
// 	spew.Dump(response, err)

	// Test GraphAPI: some logic.
	filter := "driveType eq 'project'"
	drives, _, drivesErr := client.GraphAPI.MeDrivesApi.ListMyDrives(client.GraphAPICtx).Filter(filter).Execute()
	spew.Dump(drives, drivesErr)

	drive := CollectionOfDrivesGetItem(drives, "Name", UsersMainSpace)
	if (drive == nil) {
		d := *libregraph.NewDrive(UsersMainSpace)
		d.DriveType = stringToPointer("project")
		spew.Dump(d)
		drive, _, err := client.GraphAPI.DrivesApi.CreateDrive(client.GraphAPICtx).Drive(d).Execute()
		spew.Dump(drive, err)
	}

	// Get user by name (better use mail).
	search := UserName
	users, _, usersErr := client.GraphAPI.UsersApi.ListUsers(client.GraphAPICtx).Search(search).Execute()
	spew.Dump(users, usersErr)
	userList := users.GetValue()
	if len(userList) > 0 {
		user := userList[0]
		spew.Dump(*user.Surname)

		// Create user home.
		err := client.WebdavAPI.Mkdir(fmt.Sprintf("/spaces/%s/%s", *drive.Id, *user.Surname), 0644)
		if (err == nil) {
			// Create user folders.
			values := []string{"Bill", "Contract", "Fiscal", "Legal"}
			for _, v := range values {
				err := client.WebdavAPI.Mkdir(fmt.Sprintf("/spaces/%s/%s/%s", *drive.Id, *user.Surname, v), 0644)
				if (err != nil) { spew.Dump(err) }
			}

			// Get user folders info.
			defaultProps := []string{}
			files, err := client.WebdavAPI.ReadDirWithProps(fmt.Sprintf("/spaces/%s/%s", *drive.Id, *user.Surname), defaultProps)
			spew.Dump(files, err)

			// Create recipient.
			objectID := *user.Id
			recipient := libregraph.DriveRecipient{
				ObjectId: &objectID,
			}
			recipientType := "user"
			recipient.SetLibreGraphRecipientType(recipientType)
			// Create invite.
			driveItemInvite := *libregraph.NewDriveItemInvite()
			// Set recipients (as slice of Recipient).
			driveItemInvite.SetRecipients([]libregraph.DriveRecipient{recipient})
			// Set roles.
			roleID := UnifiedRoleViewerID
			driveItemInvite.SetRoles([]string{roleID})

			for _, file := range files {
				if props, ok := file.Sys().(webdav.Props); ok {
					itemId := props.GetString(xml.Name{Space: "http://owncloud.org/ns", Local: "id"})
					spew.Dump("itemId:", itemId)

					response, _, err := client.GraphAPI.DrivesPermissionsApi.Invite(client.GraphAPICtx, *drive.Id, itemId).DriveItemInvite(driveItemInvite).Execute()
					spew.Dump(response, err)
				} else {
					spew.Dump("Props:", props, ok)
				}
			}
		} else {
			spew.Dump(err)
		}
	}


	// Test WebdavAPI: get directory content.
// 	response, err := client.WebdavAPI.ReadDir("/files/admin")
// 	spew.Dump(response, err)

}

func stringToPointer(s string) *string { return &s }

func CollectionOfDrivesGetItem(collection *libregraph.CollectionOfDrives, field, value string) *libregraph.Drive {
	for _, drive := range collection.GetValue() {
		v := reflect.ValueOf(drive)
		if v.Kind() == reflect.Struct {
			f := v.FieldByName(field)
			if f.IsValid() && f.Kind() == reflect.String && f.String() == value {
				return &drive
			}
		}
	}

	return nil
}

func CollectionOfDrivesHasItem(collection *libregraph.CollectionOfDrives, key string, value string) bool {
	for _, drive := range collection.GetValue() {
		switch key {
		case "Name":
			if drive.Name == value {
				return true
			}
		case "DriveType":
			if drive.DriveType != nil && *drive.DriveType == value {
				return true
			}
		case "Id":
			if drive.Id != nil && *drive.Id == value {
				return true
			}
		case "DriveAlias":
			if drive.DriveAlias != nil && *drive.DriveAlias == value {
				return true
			}
		default:
			// Unknown key, skip.
		}
	}

	return false
}

