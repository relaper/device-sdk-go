// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"fmt"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

func TransformWriteParameter(cv *dsModels.CommandValue, pv models.ResourceProperties) errors.EdgeX {
	if !isNumericValueType(cv) {
		return nil
	}

	value, err := commandValueForTransform(cv)
	newValue := value

	if pv.Maximum != "" {
		err = validateWriteMaximum(value, pv.Maximum)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Minimum != "" {
		err = validateWriteMinimum(value, pv.Minimum)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Offset != "" && pv.Offset != defaultOffset {
		newValue, err = transformOffset(newValue, pv.Offset, false)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Scale != "" && pv.Scale != defaultScale {
		newValue, err = transformScale(newValue, pv.Scale, false)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Base != "" && pv.Base != defaultBase {
		newValue, err = transformBase(newValue, pv.Base, false)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}

	if value != newValue {
		cv.Value = newValue
	}
	return nil
}

func validateWriteMaximum(value interface{}, maximum string) errors.EdgeX {
	switch v := value.(type) {
	case uint8:
		max, err := strconv.ParseUint(maximum, 10, 8)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最大值 %s 无法转换为 %T", maximum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v > uint8(max) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最大值", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case uint16:
		max, err := strconv.ParseUint(maximum, 10, 16)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最大值 %s 无法转换为 %T", maximum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v > uint16(max) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最大值", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case uint32:
		max, err := strconv.ParseUint(maximum, 10, 32)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最大值 %s 无法转换为 %T", maximum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v > uint32(max) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最大值", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case uint64:
		max, err := strconv.ParseUint(maximum, 10, 64)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最大值 %s 无法转换为 %T", maximum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v > max {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最大值", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int8:
		max, err := strconv.ParseInt(maximum, 10, 8)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最大值 %s 无法转换为 %T", maximum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v > int8(max) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最大值", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int16:
		max, err := strconv.ParseInt(maximum, 10, 16)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最大值 %s 无法转换为 %T", maximum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v > int16(max) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最大值", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int32:
		max, err := strconv.ParseInt(maximum, 10, 32)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最大值 %s 无法转换为 %T", maximum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v > int32(max) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最大值", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int64:
		max, err := strconv.ParseInt(maximum, 10, 64)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最大值 %s 无法转换为 %T", maximum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v > max {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最大值", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case float32:
		max, err := strconv.ParseFloat(maximum, 32)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最大值 %s 无法转换为 %T", maximum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v > float32(max) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最大值", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case float64:
		max, err := strconv.ParseFloat(maximum, 64)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最大值 %s 无法转换为 %T", maximum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v > max {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最大值", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	}
	return nil
}

func validateWriteMinimum(value interface{}, minimum string) errors.EdgeX {
	switch v := value.(type) {
	case uint8:
		min, err := strconv.ParseUint(minimum, 10, 8)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最小值 %s 无法转换为 %T", minimum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v < uint8(min) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最小值", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case uint16:
		min, err := strconv.ParseUint(minimum, 10, 16)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最小值 %s 无法转换为 %T", minimum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v < uint16(min) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最小值", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case uint32:
		min, err := strconv.ParseUint(minimum, 10, 32)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最小值 %s 无法转换为 %T", minimum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v < uint32(min) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最小值", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case uint64:
		min, err := strconv.ParseUint(minimum, 10, 64)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最小值 %s 无法转换为 %T", minimum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v < min {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最小值", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int8:
		min, err := strconv.ParseInt(minimum, 10, 8)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最小值 %s 无法转换为 %T", minimum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v < int8(min) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最小值", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int16:
		min, err := strconv.ParseInt(minimum, 10, 16)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最小值 %s 无法转换为 %T", minimum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v < int16(min) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最小值", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int32:
		min, err := strconv.ParseInt(minimum, 10, 32)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最小值 %s 无法转换为 %T", minimum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v < int32(min) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最小值", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int64:
		min, err := strconv.ParseInt(minimum, 10, 64)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最小值 %s 无法转换为 %T", minimum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v < min {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最小值", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case float32:
		min, err := strconv.ParseFloat(minimum, 32)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最小值 %s 无法转换为 %T", minimum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v < float32(min) {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最小值", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case float64:
		min, err := strconv.ParseFloat(minimum, 64)
		if err != nil {
			errMsg := fmt.Sprintf("属性配置的最小值 %s 无法转换为 %T", minimum, v)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if v < min {
			errMsg := fmt.Sprintf("设置命令参数值超出属性最小值", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	}
	return nil
}
