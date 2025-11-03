package services

import (
	"emby-client-go/internal/models"
	"crypto/md5"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// DeviceService 设备管理服务
type DeviceService struct {
	db *gorm.DB
}

// NewDeviceService 创建设备服务
func NewDeviceService(db *gorm.DB) *DeviceService {
	return &DeviceService{db: db}
}

// AddDevice 添加设备
func (d *DeviceService) AddDevice(device *models.Device) error {
	// 生成唯一标识符（如果未提供）
	if device.Identifier == "" {
		device.Identifier = d.GenerateIdentifier(device)
	}

	// 检查标识符是否已存在
	var existingDevice models.Device
	err := d.db.Where("identifier = ?", device.Identifier).First(&existingDevice).Error
	if err == nil {
		return fmt.Errorf("设备标识符已存在: %s", device.Identifier)
	}

	if err := d.db.Create(device).Error; err != nil {
		return fmt.Errorf("创建设备失败: %v", err)
	}

	log.Printf("成功添加设备: %s (%s)", device.Name, device.Identifier)
	return nil
}

// GetDevices 获取设备列表
func (d *DeviceService) GetDevices() ([]models.Device, error) {
	var devices []models.Device
	err := d.db.Preload("Servers").Find(&devices).Error
	return devices, err
}

// GetDevice 获取单个设备
func (d *DeviceService) GetDevice(id uint) (*models.Device, error) {
	var device models.Device
	err := d.db.Preload("Servers").First(&device, id).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// GetDeviceByIdentifier 根据标识符获取设备
func (d *DeviceService) GetDeviceByIdentifier(identifier string) (*models.Device, error) {
	var device models.Device
	err := d.db.Preload("Servers").Where("identifier = ?", identifier).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// UpdateDevice 更新设备
func (d *DeviceService) UpdateDevice(id uint, updates *models.Device) error {
	var device models.Device
	if err := d.db.First(&device, id).Error; err != nil {
		return err
	}

	return d.db.Model(&device).Updates(updates).Error
}

// DeleteDevice 删除设备
func (d *DeviceService) DeleteDevice(id uint) error {
	var device models.Device
	if err := d.db.First(&device, id).Error; err != nil {
		return err
	}

	// 删除关联的设备-服务器关系
	d.db.Where("device_id = ?", id).Delete(&models.DeviceServer{})

	return d.db.Delete(&device).Error
}

// UpdateLastActive 更新设备最后活跃时间
func (d *DeviceService) UpdateLastActive(id uint) error {
	now := time.Now()
	return d.db.Model(&models.Device{}).Where("id = ?", id).Update("last_active", &now).Error
}

// GenerateIdentifier 生成设备唯一标识符
func (d *DeviceService) GenerateIdentifier(device *models.Device) string {
	data := fmt.Sprintf("%s-%s-%s-%s",
		device.Name,
		device.Platform,
		device.MacAddress,
		time.Now().String())
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)[:16]
}

// AddDeviceToServer 添加设备到服务器关联
func (d *DeviceService) AddDeviceToServer(deviceID, serverID uint, priority int) error {
	var deviceServer models.DeviceServer
	err := d.db.Where("device_id = ? AND server_id = ?", deviceID, serverID).First(&deviceServer).Error
	if err == nil {
		return fmt.Errorf("设备已关联到此服务器")
	}

	deviceServer = models.DeviceServer{
		DeviceID:  deviceID,
		ServerID:  serverID,
		Priority:  priority,
		IsEnabled: true,
	}

	return d.db.Create(&deviceServer).Error
}

// RemoveDeviceFromServer 移除设备-服务器关联
func (d *DeviceService) RemoveDeviceFromServer(deviceID, serverID uint) error {
	return d.db.Where("device_id = ? AND server_id = ?", deviceID, serverID).Delete(&models.DeviceServer{}).Error
}

// GetDeviceServers 获取设备关联的服务器列表
func (d *DeviceService) GetDeviceServers(deviceID uint) ([]models.Server, error) {
	var device models.Device
	err := d.db.Preload("Servers").First(&device, deviceID).Error
	if err != nil {
		return nil, err
	}
	return device.Servers, nil
}

// SetDeviceServerPriority 设置设备-服务器关联优先级
func (d *DeviceService) SetDeviceServerPriority(deviceID, serverID uint, priority int) error {
	return d.db.Model(&models.DeviceServer{}).
		Where("device_id = ? AND server_id = ?", deviceID, serverID).
		Update("priority", priority).Error
}

// EnableDeviceServer 启用设备-服务器关联
func (d *DeviceService) EnableDeviceServer(deviceID, serverID uint, enabled bool) error {
	return d.db.Model(&models.DeviceServer{}).
		Where("device_id = ? AND server_id = ?", deviceID, serverID).
		Update("is_enabled", enabled).Error
}

// UpdateDeviceLastUsed 更新设备在指定服务器的最后使用时间
func (d *DeviceService) UpdateDeviceLastUsed(deviceID, serverID uint) error {
	now := time.Now()
	return d.db.Model(&models.DeviceServer{}).
		Where("device_id = ? AND server_id = ?", deviceID, serverID).
		Update("last_used", &now).Error
}

// GetDevicesByServer 获取指定服务器关联的所有设备
func (d *DeviceService) GetDevicesByServer(serverID uint) ([]models.Device, error) {
	var devices []models.Device
	err := d.db.Joins("JOIN device_servers ON devices.id = device_servers.device_id").
		Where("device_servers.server_id = ?", serverID).
		Preload("Servers").
		Find(&devices).Error
	return devices, err
}

// GetActiveDevices 获取活跃设备列表（最近30天有活动）
func (d *DeviceService) GetActiveDevices() ([]models.Device, error) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var devices []models.Device
	err := d.db.Where("last_active >= ?", thirtyDaysAgo).
		Or("last_active IS NULL").
		Preload("Servers").
		Find(&devices).Error
	return devices, err
}

// GetInactiveDevices 获取非活跃设备列表
func (d *DeviceService) GetInactiveDevices() ([]models.Device, error) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var devices []models.Device
	err := d.db.Where("last_active < ?", thirtyDaysAgo).
		Preload("Servers").
		Find(&devices).Error
	return devices, err
}