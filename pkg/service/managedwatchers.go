// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
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

// AddProvisionWatcher adds a new Watcher to the cache and Core Metadata
// Returns new Watcher id or non-nil error.
func (s *DeviceService) AddProvisionWatcher(watcher models.ProvisionWatcher) (string, error) {
	if pw, ok := cache.ProvisionWatchers().ForName(watcher.Name); ok {
		return pw.Id,
			errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("ProvisionWatcher %s 已存在", watcher.Name), nil)
	}

	_, ok := cache.Profiles().ForName(watcher.ProfileName)
	if !ok {
		res, err := s.edgexClients.DeviceProfileClient.DeviceProfileByName(context.Background(), watcher.ProfileName)
		if err != nil {
			errMsg := fmt.Sprintf("查找 provision watcher %s 的模型 %s 失败", watcher.Name, watcher.ProfileName)
			s.LoggingClient.Error(errMsg)
			return "", err
		}
		err = cache.Profiles().Add(dtos.ToDeviceProfileModel(res.Profile))
		if err != nil {
			return "", errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("缓存模型 %s 失败", res.Profile.Name), err)
		}
	}
	watcher.ServiceName = s.ServiceName

	s.LoggingClient.Debugf("Adding managed ProvisionWatcher %s", watcher.Name)
	req := requests.NewAddProvisionWatcherRequest(dtos.FromProvisionWatcherModelToDTO(watcher))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	res, err := s.edgexClients.ProvisionWatcherClient.Add(ctx, []requests.AddProvisionWatcherRequest{req})
	if err != nil {
		s.LoggingClient.Errorf("核心元数据服务添加 ProvisionWatcher %s 失败: %v", watcher.Name, err)
		return "", err
	}

	return res[0].Id, nil
}

// ProvisionWatchers return all managed Watchers from cache
func (s *DeviceService) ProvisionWatchers() []models.ProvisionWatcher {
	return cache.ProvisionWatchers().All()
}

// GetProvisionWatcherByName returns the Watcher by its name if it exists in the cache, or returns an error.
func (s *DeviceService) GetProvisionWatcherByName(name string) (models.ProvisionWatcher, error) {
	pw, ok := cache.ProvisionWatchers().ForName(name)
	if !ok {
		msg := fmt.Sprintf("查找ProvisionWatcher %s 缓存失败", name)
		s.LoggingClient.Error(msg)
		return models.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}
	return pw, nil
}

// RemoveProvisionWatcher removes the specified Watcher by name from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *DeviceService) RemoveProvisionWatcher(name string) error {
	pw, ok := cache.ProvisionWatchers().ForName(name)
	if !ok {
		msg := fmt.Sprintf("查找ProvisionWatcher %s 缓存失败", name)
		s.LoggingClient.Error(msg)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}

	s.LoggingClient.Debugf("Removing managed ProvisionWatcher: %s", pw.Name)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	_, err := s.edgexClients.ProvisionWatcherClient.DeleteProvisionWatcherByName(ctx, name)
	if err != nil {
		s.LoggingClient.Errorf("核心元服务删除 ProvisionWatcher %s 失败", name)
		return err
	}

	return nil
}

// UpdateProvisionWatcher updates the Watcher in the cache and ensures that the
// copy in Core Metadata is also updated.
func (s *DeviceService) UpdateProvisionWatcher(watcher models.ProvisionWatcher) error {
	_, ok := cache.ProvisionWatchers().ForName(watcher.Name)
	if !ok {
		msg := fmt.Sprintf("查找ProvisionWatcher %s 缓存失败", watcher.Name)
		s.LoggingClient.Error(msg)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}

	s.LoggingClient.Debugf("Updating managed ProvisionWatcher: %s", watcher.Name)
	req := requests.NewUpdateProvisionWatcherRequest(dtos.FromProvisionWatcherModelToUpdateDTO(watcher))
	req.ProvisionWatcher.Id = nil
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	_, err := s.edgexClients.ProvisionWatcherClient.Update(ctx, []requests.UpdateProvisionWatcherRequest{req})
	if err != nil {
		s.LoggingClient.Errorf("核心元数据服务更新 ProvisionWatcher %s 失败： %v", watcher.Name, err)
		return err
	}

	return nil
}
