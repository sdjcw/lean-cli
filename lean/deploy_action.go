package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/codegangsta/cli"
	"github.com/jhoonb/archivex"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
)

func determineGroupName(appID string) (string, error) {
	info, err := api.GetAppInfo(appID)
	if err != nil {
		return "", err
	}
	mode := info.LeanEngineMode

	groups, err := api.GetGroups(appID)
	if err != nil {
		return "", err
	}

	groupName, found, err := linq.From(groups).Where(func(group linq.T) (bool, error) {
		groupName := group.(*api.GetGroupsResult).GroupName
		if mode == "free" {
			return groupName != "staging", nil
		}
		return groupName == "staging", nil
	}).Select(func(group linq.T) (linq.T, error) {
		return group.(*api.GetGroupsResult).GroupName, nil
	}).First()
	if err != nil {
		return "", err
	}
	if !found {
		return "", errors.New("group not found")
	}
	return groupName.(string), nil
}

func uploadProject(appID string, repoPath string) (*api.UploadFileResult, error) {
	// TODO: ignore files

	fileDir, err := ioutil.TempDir("", "leanengine")
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(fileDir, "leanengine.zip")

	op.Write("压缩项目文件 ...")
	zip := new(archivex.ZipFile)
	func() {
		defer zip.Close()
		zip.Create(filePath)
		zip.AddAll(repoPath, false)
	}()
	op.Successed()

	file, err := api.UploadFile(appID, filePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func deployFromLocal(appID string, groupName string) error {
	file, err := uploadProject(appID, "")
	if err != nil {
		return err
	}

	defer func() {
		op.Write("删除临时文件")
		err := api.DeleteFile(appID, file.ObjectID)
		if err != nil {
			op.Failed()
		} else {
			op.Successed()
		}
	}()

	eventTok, err := api.DeployAppFromFile("", groupName, file.URL)
	ok, err := api.PollEvents(appID, eventTok, os.Stdout)
	if err != nil {
		return err
	}
	if !ok {
		return cli.NewExitError("部署失败", 1)
	}
	return nil
}

func deployFromGit(appID string, groupName string) error {
	eventTok, err := api.DeployAppFromGit("", groupName)
	if err != nil {
		return err
	}
	ok, err := api.PollEvents(appID, eventTok, os.Stdout)
	if err != nil {
		return err
	}
	if !ok {
		return cli.NewExitError("部署失败", 1)
	}
	return nil
}

func deployAction(*cli.Context) error {
	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		log.Fatalln("没有关联任何 app，请使用 lean checkout 来关联应用。")
	}
	if err != nil {
		return newCliError(err)
	}

	op.Write("获取部署分组信息")
	groupName, err := determineGroupName(appID)
	if err != nil {
		op.Failed()
		return newCliError(err)
	}
	op.Successed()

	if groupName == "staging" {
		op.Write("准备部署应用到预备环境")
	} else {
		op.Write("准备部署应用到生产环境: " + groupName)
	}

	if isDeployFromGit {
		err = deployFromGit(appID, groupName)
		if err != nil {
			return newCliError(err)
		}
	} else {
		err = deployFromLocal(appID, groupName)
		if err != nil {
			return newCliError(err)
		}
	}
	return nil
}
