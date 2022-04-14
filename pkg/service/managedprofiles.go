// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/google/uuid"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/cache"
)

// AddDeviceProfile adds a new DeviceProfile to the Device Service and Core Metadata
// Returns new DeviceProfile id or non-nil error.
func (s *DeviceService) AddDeviceProfile(profile models.DeviceProfile) (string, error) {
	if p, ok := cache.Profiles().ForName(profile.Name); ok {
		return p.Id, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("模型名称 %s 已存在", profile.Name), nil)
	}

	s.LoggingClient.Debugf("Adding managed Profile %s", profile.Name)
	req := requests.NewDeviceProfileRequest(dtos.FromDeviceProfileModelToDTO(profile))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	res, err := s.edgexClients.DeviceProfileClient.Add(ctx, []requests.DeviceProfileRequest{req})
	if err != nil {
		s.LoggingClient.Errorf("failed to add Profile %s to Core Metadata: %v", profile.Name, err)
		return "", err
	}
	err = cache.Profiles().Add(profile)
	if err != nil {
		s.LoggingClient.Errorf("failed to add Profile %s to cache", profile.Name)
		return "", err
	}

	return res[0].Id, nil
}

// DeviceProfiles return all managed DeviceProfiles from cache
func (s *DeviceService) DeviceProfiles() []models.DeviceProfile {
	return cache.Profiles().All()
}

// GetProfileByName returns the Profile by its name if it exists in the cache, or returns an error.
func (s *DeviceService) GetProfileByName(name string) (models.DeviceProfile, error) {
	profile, ok := cache.Profiles().ForName(name)
	if !ok {
		msg := fmt.Sprintf("查找模型缓存 %s 失败", name)
		s.LoggingClient.Error(msg)
		return models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}
	return profile, nil
}

// RemoveDeviceProfileByName removes the specified DeviceProfile by name from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *DeviceService) RemoveDeviceProfileByName(name string) error {
	profile, ok := cache.Profiles().ForName(name)
	if !ok {
		msg := fmt.Sprintf("查找模型缓存 %s 失", name)
		s.LoggingClient.Error(msg)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}

	s.LoggingClient.Debugf("Removing managed Profile %s", profile.Name)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	_, err := s.edgexClients.DeviceProfileClient.DeleteByName(ctx, name)
	if err != nil {
		s.LoggingClient.Errorf("核心元数据服务删除模型 %s 失败", name)
		return err
	}

	err = cache.Profiles().RemoveByName(name)
	return err
}

// UpdateDeviceProfile updates the DeviceProfile in the cache and ensures that the
// copy in Core Metadata is also updated.
func (s *DeviceService) UpdateDeviceProfile(profile models.DeviceProfile) error {
	_, ok := cache.Profiles().ForName(profile.Name)
	if !ok {
		msg := fmt.Sprintf("查找模型缓存 %s 失", profile.Name)
		s.LoggingClient.Error(msg)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}

	s.LoggingClient.Debugf("Updating managed Profile %s", profile.Name)
	req := requests.NewDeviceProfileRequest(dtos.FromDeviceProfileModelToDTO(profile))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	_, err := s.edgexClients.DeviceProfileClient.Update(ctx, []requests.DeviceProfileRequest{req})
	if err != nil {
		s.LoggingClient.Errorf("核心元数据服务 %s 更新模型失败: %v", profile.Name, err)
		return err
	}

	return err
}

// DeviceCommand retrieves the specific DeviceCommand instance from cache according to
// the Device name and Command name
func (s *DeviceService) DeviceCommand(deviceName string, commandName string) (models.DeviceCommand, bool) {
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		s.LoggingClient.Errorf("查找设备缓存 %s 失", deviceName)
		return models.DeviceCommand{}, false
	}

	dc, ok := cache.Profiles().DeviceCommand(device.ProfileName, commandName)
	if !ok {
		return dc, false
	}
	return dc, true
}

// DeviceResource retrieves the specific DeviceResource instance from cache according to
// the Device name and Device Resource name
func (s *DeviceService) DeviceResource(deviceName string, deviceResource string) (models.DeviceResource, bool) {
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		s.LoggingClient.Errorf("查找设备缓存 %s 失", deviceName)
		return models.DeviceResource{}, false
	}

	dr, ok := cache.Profiles().DeviceResource(device.ProfileName, deviceResource)
	if !ok {
		return dr, false
	}
	return dr, true
}
