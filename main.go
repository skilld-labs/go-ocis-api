package main

import (
	"fmt"
	"reflect"
	"encoding/xml"
	"strconv"
// 	"net/url"

	"main/docs"
	"main/rest"
	"github.com/davecgh/go-spew/spew"
	libregraph "github.com/owncloud/libre-graph-api-go"
	webdav "github.com/studio-b12/gowebdav"
)

type provider struct {
	client *rest.Client
}

func main() {

// 	uri, err := url.Parse("https://aaaaa:bbbbb@my.test.com")
// 	spew.Dump(uri.Port(), err)
// 	return

	cfg := docs.Config{
		Host: "***",
		Username: "***",
		Password: "***",
	}

	client := docs.New(cfg)
	p := &provider{client: client}

	// Test GraphAPI: get current user.
// 	response, _, err := client.GraphAPI.MeUserApi.GetOwnUser(client.GraphAPICtx).Execute()
// 	spew.Dump(response, err)

	// Test GraphAPI: get all drives where the current user is a regular member of.
// 	filter := "driveType eq 'project'"
// 	response, _, err := client.GraphAPI.MeDrivesApi.ListMyDrives(client.GraphAPICtx).Filter(filter).Execute()
// 	spew.Dump(response, err)

	// Test GraphAPI: Create a new drive of a specific type.
// 	drive := *libregraph.NewDrive(UsersSpace)
// 	drive.DriveType = stringToPointer("project")
// 	spew.Dump(drive)
// 	response, _, err := client.GraphAPI.DrivesApi.CreateDrive(client.GraphAPICtx).Drive(drive).Execute()
// 	spew.Dump(response, err)

	// Test GraphAPI: some logic.
// 	drives, drivesErr := p.listMyDrives("project")
// 	spew.Dump(drives, drivesErr)
// 	drive := CollectionOfDrivesGetItem(drives, "Name", UsersSpace)

// 	drive := p.myDrivesGetItemBy("Name", UsersSpace)
// 	spew.Dump(drive)
// 	if (drive == nil) {
// 		d := *libregraph.NewDrive(UsersSpace)
// 		d.DriveType = stringToPointer("project")
// 		spew.Dump(d)
// 		drive, _, err := client.GraphAPI.DrivesApi.CreateDrive(client.GraphAPICtx).Drive(d).Execute()
// 		spew.Dump(drive, err)
// 	}

	drive := p.myDrivesGetOrCreateItem(UsersSpace)
	spew.Dump(drive)

	// Get user by name (better use mail).
	user := p.searchUser(UserMail)
	spew.Dump(user)
	if user != nil {
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
			driveItemInvite := CreateDriveItemInvite(*user.Id, "user", UnifiedRoleViewerID)
			spew.Dump(driveItemInvite)

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

func CreateDriveItemInvite(objectID, recipientType, roleID string) libregraph.DriveItemInvite {
	// Create recipient.
	recipient := libregraph.DriveRecipient{
		ObjectId: &objectID,
	}
	recipient.SetLibreGraphRecipientType(recipientType)
	// Create invite.
	driveItemInvite := *libregraph.NewDriveItemInvite()
	// Set recipients (as slice of Recipient).
	driveItemInvite.SetRecipients([]libregraph.DriveRecipient{recipient})
	// Set roles.
	driveItemInvite.SetRoles([]string{roleID})

	return driveItemInvite
}

func (p *provider) searchUser(search string) *libregraph.User {
	users := p.getUsersList(search)
	spew.Dump(users)
	userList := users.GetValue()
	if len(userList) > 0 {
		return &userList[0]
	}

	return nil
}

func (p *provider) getUsersList(search string) *libregraph.CollectionOfUser {
	users, _, _ := p.client.GraphAPI.UsersApi.ListUsers(p.client.GraphAPICtx).Search(strconv.Quote(search)).Execute()

	return users
}

func (p *provider) myDrivesGetOrCreateItem(value string) *libregraph.Drive {
	drive := p.myDrivesGetItemBy("Name", value)
	if (drive == nil) {
		drive = p.myDrivesCreateDrive(value, "project")
	}

	return drive
}

func (p *provider) myDrivesCreateDrive(name, driveType string) *libregraph.Drive {
	d := *libregraph.NewDrive(name)
	d.DriveType = &driveType
	drive, _, _ := p.client.GraphAPI.DrivesApi.CreateDrive(p.client.GraphAPICtx).Drive(d).Execute()

	return drive
}

func (p *provider) myDrivesGetItemBy(field, value string) *libregraph.Drive {
	if drives, err := p.listMyDrives("project"); err == nil {
		return CollectionOfDrivesGetItemBy(drives, field, value)
	}

	return nil
}

func (p *provider) listMyDrives(driveType string) (*libregraph.CollectionOfDrives, error) {
	filter := ""
	if driveType != "" {
		filter = fmt.Sprintf("driveType eq '%s'", driveType)
	}
	drives, _, err := p.client.GraphAPI.MeDrivesApi.ListMyDrives(p.client.GraphAPICtx).Filter(filter).Execute()

	return drives, err
}

func CollectionOfDrivesGetItemBy(collection *libregraph.CollectionOfDrives, field, value string) *libregraph.Drive {
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

func stringToPointer(s string) *string { return &s }

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

