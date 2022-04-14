// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
)

func UpdateProfile(profileRequest requests.DeviceProfileRequest, lc logger.LoggingClient) errors.EdgeX {
	_, ok := cache.Profiles().ForName(profileRequest.Profile.Name)
	if !ok {
		errMsg := fmt.Sprintf("查找模型失败 %s", profileRequest.Profile.Name)
		return errors.NewCommonEdgeX(errors.KindInvalidId, errMsg, nil)
	}

	err := cache.Profiles().Update(dtos.ToDeviceProfileModel(profileRequest.Profile))
	if err != nil {
		errMsg := fmt.Sprintf("更新模型失败 %s", profileRequest.Profile.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	lc.Debug(fmt.Sprintf("profile %s updated", profileRequest.Profile.Name))
	return nil
}

func AddDevice(addDeviceRequest requests.AddDeviceRequest, dic *di.Container) errors.EdgeX {
	device := dtos.ToDeviceModel(addDeviceRequest.Device)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	edgexErr := updateAssociatedProfile(device.ProfileName, dic)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	edgexErr = cache.Devices().Add(device)
	if edgexErr != nil {
		errMsg := fmt.Sprintf("添加模型 %s 失败", device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, edgexErr)
	}
	lc.Debugf("device %s added", device.Name)

	driver := container.ProtocolDriverFrom(dic.Get)
	err := driver.AddDevice(device.Name, device.Protocols, device.AdminState)
	if err == nil {
		lc.Debugf("Invoked driver.AddDevice callback for %s", device.Name)
	} else {
		errMsg := fmt.Sprintf("添加设备（%s）回调失败", device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	lc.Debugf("starting AutoEvents for device %s", device.Name)
	container.ManagerFrom(dic.Get).RestartForDevice(device.Name)
	return nil
}

func UpdateDevice(updateDeviceRequest requests.UpdateDeviceRequest, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	ds := container.DeviceServiceFrom(dic.Get)

	device, exist := cache.Devices().ForName(*updateDeviceRequest.Device.Name)
	if !exist {
		// scenario that device migrate from another device service to here
		if ds.Name == *updateDeviceRequest.Device.ServiceName {
			var newDevice models.Device
			requests.ReplaceDeviceModelFieldsWithDTO(&newDevice, updateDeviceRequest.Device)
			req := requests.NewAddDeviceRequest(dtos.FromDeviceModelToDTO(newDevice))
			return AddDevice(req, dic)
		} else {
			errMsg := fmt.Sprintf("查找设备 %s 失败", *updateDeviceRequest.Device.ServiceName)
			return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
		}
	}
	// scenario that device is moving to another device service
	if ds.Name != *updateDeviceRequest.Device.ServiceName {
		return DeleteDevice(*updateDeviceRequest.Device.Name, dic)
	}

	requests.ReplaceDeviceModelFieldsWithDTO(&device, updateDeviceRequest.Device)
	edgexErr := updateAssociatedProfile(device.ProfileName, dic)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	edgexErr = cache.Devices().Update(device)
	if edgexErr != nil {
		errMsg := fmt.Sprintf("更新设备 %s 失败", device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, edgexErr)
	}
	lc.Debugf("device %s updated", device.Name)

	driver := container.ProtocolDriverFrom(dic.Get)
	err := driver.UpdateDevice(device.Name, device.Protocols, device.AdminState)
	if err == nil {
		lc.Debugf("Invoked driver.UpdateDevice callback for %s", device.Name)
	} else {
		errMsg := fmt.Sprintf("更新设备（%s）回调失败", device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	autoEventManager := container.ManagerFrom(dic.Get)
	if device.AdminState == models.Locked {
		lc.Debugf("stopping AutoEvents for the locked device %s", device.Name)
		autoEventManager.StopForDevice(device.Name)
	} else {
		lc.Debugf("starting AutoEvents for device %s", device.Name)
		autoEventManager.RestartForDevice(device.Name)
	}
	return nil
}

func DeleteDevice(name string, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	// check the device exist and stop its autoevents
	device, ok := cache.Devices().ForName(name)
	if ok {
		lc.Debugf("stopping AutoEvents for device %s", device.Name)
		container.ManagerFrom(dic.Get).StopForDevice(device.Name)
	} else {
		errMsg := fmt.Sprintf("查找设备 %s 失败", name)
		return errors.NewCommonEdgeX(errors.KindInvalidId, errMsg, nil)
	}

	// remove the device in cache
	edgexErr := cache.Devices().RemoveByName(name)
	if edgexErr != nil {
		errMsg := fmt.Sprintf("删除设备 %s 失败", device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, edgexErr)
	}
	lc.Debugf("Removed device: %s", device.Name)

	driver := container.ProtocolDriverFrom(dic.Get)
	err := driver.RemoveDevice(device.Name, device.Protocols)
	if err == nil {
		lc.Debugf("Invoked driver.RemoveDevice callback for %s", device.Name)
	} else {
		errMsg := fmt.Sprintf("删除设备（%s）回调失败", device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	// a special case in which user updates the device profile after deleting all
	// devices in metadata, the profile won't be correctly updated because metadata
	// does not know which device service callback it needs to call. Remove the unused
	// device profile in cache so that if it is updated in metadata, next time the
	// device using it is added/updated, the cache can receive the updated one as well.
	if cache.CheckProfileNotUsed(device.ProfileName) {
		edgexErr = cache.Profiles().RemoveByName(device.ProfileName)
		if edgexErr != nil {
			lc.Warn("failed to remove unused profile", edgexErr.DebugMessages())
		}
	}

	return nil
}

func AddProvisionWatcher(addProvisionWatcherRequest requests.AddProvisionWatcherRequest, lc logger.LoggingClient, dic *di.Container) errors.EdgeX {
	provisionWatcher := dtos.ToProvisionWatcherModel(addProvisionWatcherRequest.ProvisionWatcher)

	edgexErr := updateAssociatedProfile(provisionWatcher.ProfileName, dic)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	edgexErr = cache.ProvisionWatchers().Add(provisionWatcher)
	if edgexErr != nil {
		errMsg := fmt.Sprintf("添加 provision watcher %s 失败", provisionWatcher.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, edgexErr)
	}

	lc.Debugf("provision watcher %s added", provisionWatcher.Name)
	return nil
}

func UpdateProvisionWatcher(updateProvisionWatcherRequest requests.UpdateProvisionWatcherRequest, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	ds := container.DeviceServiceFrom(dic.Get)

	provisionWatcher, exist := cache.ProvisionWatchers().ForName(*updateProvisionWatcherRequest.ProvisionWatcher.Name)
	if !exist {
		if ds.Name == *updateProvisionWatcherRequest.ProvisionWatcher.ServiceName {
			var newProvisionWatcher models.ProvisionWatcher
			requests.ReplaceProvisionWatcherModelFieldsWithDTO(&newProvisionWatcher, updateProvisionWatcherRequest.ProvisionWatcher)
			req := requests.NewAddProvisionWatcherRequest(dtos.FromProvisionWatcherModelToDTO(newProvisionWatcher))
			return AddProvisionWatcher(req, lc, dic)
		} else {
			errMsg := fmt.Sprintf("查找 provision watcher %s 失败", *updateProvisionWatcherRequest.ProvisionWatcher.ServiceName)
			return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
		}
	}
	if ds.Name != *updateProvisionWatcherRequest.ProvisionWatcher.ServiceName {
		return DeleteProvisionWatcher(*updateProvisionWatcherRequest.ProvisionWatcher.Name, lc)
	}

	requests.ReplaceProvisionWatcherModelFieldsWithDTO(&provisionWatcher, updateProvisionWatcherRequest.ProvisionWatcher)
	edgexErr := updateAssociatedProfile(provisionWatcher.ProfileName, dic)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	edgexErr = cache.ProvisionWatchers().Update(provisionWatcher)
	if edgexErr != nil {
		errMsg := fmt.Sprintf("更新 provision watcher %s 失败", provisionWatcher.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, edgexErr)
	}

	lc.Debugf("provision watcher %s updated", provisionWatcher.Name)
	return nil
}

func DeleteProvisionWatcher(name string, lc logger.LoggingClient) errors.EdgeX {
	err := cache.ProvisionWatchers().RemoveByName(name)
	if err != nil {
		errMsg := fmt.Sprintf("删除 provision watcher %s 失败", name)
		return errors.NewCommonEdgeX(errors.KindInvalidId, errMsg, err)
	}

	lc.Debugf("removed provision watcher %s", name)
	return nil
}

func UpdateDeviceService(updateDeviceServiceRequest requests.UpdateDeviceServiceRequest, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	ds := container.DeviceServiceFrom(dic.Get)

	// we use request.Service.Name to identify the device service (i.e. it's not updatable)
	// so if the request's service name is inconsistent with device service name
	// we should not update it.
	if ds.Name != *updateDeviceServiceRequest.Service.Name {
		errMsg := fmt.Sprintf("鉴别驱动 %s 失败", *updateDeviceServiceRequest.Service.Name)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}

	ds.AdminState = models.AdminState(*updateDeviceServiceRequest.Service.AdminState)
	ds.Labels = updateDeviceServiceRequest.Service.Labels

	lc.Debug("device service updated")
	return nil
}

// updateAssociatedProfile updates the profile specified in AddDeviceRequest or UpdateDeviceRequest or AddProvisionWatcherRequest or UpdateProvisionWatcherRequest
// to stay consistent with core metadata.
func updateAssociatedProfile(profileName string, dic *di.Container) errors.EdgeX {
	dpc := bootstrapContainer.MetadataDeviceProfileClientFrom(dic.Get)

	res, err := dpc.DeviceProfileByName(context.Background(), profileName)
	if err != nil {
		errMsg := fmt.Sprintf("从元数据服务获取模型 %s 失败", profileName)
		return errors.NewCommonEdgeX(errors.KindInvalidId, errMsg, err)
	}

	_, exist := cache.Profiles().ForName(profileName)
	if !exist {
		err = cache.Profiles().Add(dtos.ToDeviceProfileModel(res.Profile))
		if err != nil {
			errMsg := fmt.Sprintf("添加模型 %s 失败", profileName)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		return nil
	}
	err = cache.Profiles().Update(dtos.ToDeviceProfileModel(res.Profile))
	if err != nil {
		errMsg := fmt.Sprintf("在缓存中更新模型 %s 失败，使用旧数据", profileName)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	return nil
}
