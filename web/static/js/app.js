// 全局变量
let authToken = null;
let currentUser = null;

// API 基础URL
const API_BASE = '/api';

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    checkAuth();
    setupEventListeners();
});

// 检查认证状态
function checkAuth() {
    authToken = localStorage.getItem('authToken');
    if (!authToken) {
        showLoginModal();
    } else {
        loadUserProfile();
        showMainContent();
    }
}

// 设置事件监听器
function setupEventListeners() {
    // 登录表单
    document.getElementById('loginForm').addEventListener('submit', function(e) {
        e.preventDefault();
        login();
    });

    // 导航标签
    document.querySelectorAll('[data-tab]').forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            switchTab(this.dataset.tab);
        });
    });

    // 退出登录
    document.getElementById('logout-btn').addEventListener('click', logout);

    // 添加设备表单
    document.getElementById('addDeviceForm').addEventListener('submit', function(e) {
        e.preventDefault();
        addDevice();
    });

    // 添加服务器表单
    document.getElementById('addServerForm').addEventListener('submit', function(e) {
        e.preventDefault();
        addServer();
    });
}

// 登录
async function login() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password })
        });

        if (!response.ok) {
            throw new Error('登录失败');
        }

        const data = await response.json();
        authToken = data.token;
        currentUser = data.user;

        localStorage.setItem('authToken', authToken);

        // 关闭登录模态框
        const modal = bootstrap.Modal.getInstance(document.getElementById('loginModal'));
        modal.hide();

        showMainContent();
        loadDashboard();
    } catch (error) {
        showAlert('登录失败，请检查用户名和密码', 'danger');
    }
}

// 退出登录
function logout() {
    localStorage.removeItem('authToken');
    authToken = null;
    currentUser = null;
    location.reload();
}

// 加载用户信息
async function loadUserProfile() {
    try {
        const response = await apiCall(`${API_BASE}/auth/me`);
        currentUser = response;
    } catch (error) {
        logout();
    }
}

// 显示主内容
function showMainContent() {
    document.getElementById('loginModal').classList.remove('show');
    loadDashboard();
}

// 切换标签页
function switchTab(tabName) {
    // 更新导航
    document.querySelectorAll('.list-group-item').forEach(item => {
        item.classList.remove('active');
    });
    document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

    // 更新内容
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.add('hidden');
    });
    document.getElementById(`${tabName}-content`).classList.remove('hidden');

    // 加载对应数据
    switch(tabName) {
        case 'dashboard':
            loadDashboard();
            break;
        case 'devices':
            loadDevices();
            break;
        case 'servers':
            loadServers();
            break;
        case 'associations':
            loadAssociations();
            break;
    }
}

// API 调用封装
async function apiCall(url, options = {}) {
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${authToken}`
        }
    };

    const response = await fetch(url, { ...defaultOptions, ...options });

    if (response.status === 401) {
        logout();
        return;
    }

    if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
    }

    return response.json();
}

// 显示提示消息
function showAlert(message, type = 'info') {
    const alertHtml = `
        <div class="alert alert-${type} alert-dismissible fade show" role="alert">
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        </div>
    `;

    const container = document.querySelector('.container-fluid');
    container.insertAdjacentHTML('afterbegin', alertHtml);

    // 3秒后自动关闭
    setTimeout(() => {
        const alert = container.querySelector('.alert');
        if (alert) {
            const bsAlert = bootstrap.Alert.getInstance(alert);
            if (bsAlert) bsAlert.close();
        }
    }, 3000);
}

// 加载仪表板
async function loadDashboard() {
    try {
        const [devices, servers] = await Promise.all([
            apiCall(`${API_BASE}/devices`),
            apiCall(`${API_BASE}/servers`)
        ]);

        // 更新统计数据
        document.getElementById('total-devices').textContent = devices.data.length;
        document.getElementById('total-servers').textContent = servers.data.length;

        const activeDevices = devices.data.filter(device => {
            if (!device.last_active) return true;
            return new Date(device.last_active) > new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
        });
        document.getElementById('active-devices').textContent = activeDevices.length;

        // 这里应该检查服务器状态，暂时显示总数
        document.getElementById('online-servers').textContent = servers.data.length;

    } catch (error) {
        console.error('加载仪表板失败:', error);
    }
}

// 加载设备列表
async function loadDevices() {
    try {
        const response = await apiCall(`${API_BASE}/devices`);
        const devices = response.data;

        const tbody = document.getElementById('devices-list');
        tbody.innerHTML = '';

        devices.forEach(device => {
            const lastActive = device.last_active ?
                new Date(device.last_active).toLocaleString() : '从未';
            const status = device.is_active ?
                '<span class="badge bg-success">活跃</span>' :
                '<span class="badge bg-secondary">非活跃</span>';

            const row = `
                <tr>
                    <td>${device.name}</td>
                    <td><code>${device.identifier}</code></td>
                    <td>${device.platform || '-'}</td>
                    <td>${device.ip_address || '-'}</td>
                    <td>${lastActive}</td>
                    <td>${status}</td>
                    <td>
                        <button class="btn btn-sm btn-outline-primary" onclick="editDevice(${device.id})">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="btn btn-sm btn-outline-danger" onclick="deleteDevice(${device.id})">
                            <i class="fas fa-trash"></i>
                        </button>
                    </td>
                </tr>
            `;
            tbody.innerHTML += row;
        });

        if (devices.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" class="text-center">暂无设备</td></tr>';
        }
    } catch (error) {
        showAlert('加载设备列表失败', 'danger');
    }
}

// 加载服务器列表
async function loadServers() {
    try {
        const response = await apiCall(`${API_BASE}/servers`);
        const servers = response.data;

        const tbody = document.getElementById('servers-list');
        tbody.innerHTML = '';

        servers.forEach(server => {
            const lastCheck = server.last_check ?
                new Date(server.last_check).toLocaleString() : '从未';
            const status = server.is_active ?
                '<span class="badge bg-success">在线</span>' :
                '<span class="badge bg-danger">离线</span>';

            const row = `
                <tr>
                    <td>${server.name}</td>
                    <td><a href="${server.url}" target="_blank">${server.url}</a></td>
                    <td>${server.version || '-'}</td>
                    <td>${lastCheck}</td>
                    <td>${status}</td>
                    <td>
                        <button class="btn btn-sm btn-outline-success" onclick="testServer(${server.id})">
                            <i class="fas fa-plug"></i> 测试
                        </button>
                        <button class="btn btn-sm btn-outline-info" onclick="syncDevices(${server.id})">
                            <i class="fas fa-sync"></i> 同步
                        </button>
                        <button class="btn btn-sm btn-outline-primary" onclick="editServer(${server.id})">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="btn btn-sm btn-outline-danger" onclick="deleteServer(${server.id})">
                            <i class="fas fa-trash"></i>
                        </button>
                    </td>
                </tr>
            `;
            tbody.innerHTML += row;
        });

        if (servers.length === 0) {
            tbody.innerHTML = '<tr><td colspan="6" class="text-center">暂无服务器</td></tr>';
        }
    } catch (error) {
        showAlert('加载服务器列表失败', 'danger');
    }
}

// 加载关联关系
async function loadAssociations() {
    try {
        const [devices, servers] = await Promise.all([
            apiCall(`${API_BASE}/devices`),
            apiCall(`${API_BASE}/servers`)
        ]);

        // 更新下拉选择框
        const deviceSelect = document.getElementById('select-device');
        const serverSelect = document.getElementById('select-server');

        deviceSelect.innerHTML = '<option value="">选择设备</option>';
        devices.data.forEach(device => {
            deviceSelect.innerHTML += `<option value="${device.id}">${device.name}</option>`;
        });

        serverSelect.innerHTML = '<option value="">选择服务器</option>';
        servers.data.forEach(server => {
            serverSelect.innerHTML += `<option value="${server.id}">${server.name}</option>`;
        });

        // 加载现有关联关系
        const associationsList = document.getElementById('associations-list');
        associationsList.innerHTML = '<div class="text-center">选择设备查看关联关系...</div>';
    } catch (error) {
        showAlert('加载关联关系失败', 'danger');
    }
}

// 添加设备
async function addDevice() {
    const deviceData = {
        name: document.getElementById('device-name').value,
        platform: document.getElementById('device-platform').value,
        ip_address: document.getElementById('device-ip').value,
        mac_address: document.getElementById('device-mac').value,
        description: document.getElementById('device-desc').value
    };

    try {
        await apiCall(`${API_BASE}/devices`, {
            method: 'POST',
            body: JSON.stringify(deviceData)
        });

        showAlert('设备添加成功', 'success');

        // 重置表单
        document.getElementById('addDeviceForm').reset();

        // 关闭模态框
        const modal = bootstrap.Modal.getInstance(document.getElementById('addDeviceModal'));
        modal.hide();

        // 刷新设备列表
        loadDevices();
    } catch (error) {
        showAlert('添加设备失败', 'danger');
    }
}

// 添加服务器
async function addServer() {
    const serverData = {
        name: document.getElementById('server-name').value,
        url: document.getElementById('server-url').value,
        user_name: document.getElementById('server-username').value,
        password: document.getElementById('server-password').value,
        description: document.getElementById('server_desc').value
    };

    try {
        await apiCall(`${API_BASE}/servers`, {
            method: 'POST',
            body: JSON.stringify(serverData)
        });

        showAlert('服务器添加成功', 'success');

        // 重置表单
        document.getElementById('addServerForm').reset();

        // 关闭模态框
        const modal = bootstrap.Modal.getInstance(document.getElementById('addServerModal'));
        modal.hide();

        // 刷新服务器列表
        loadServers();
    } catch (error) {
        showAlert('添加服务器失败: ' + error.message, 'danger');
    }
}

// 测试服务器连接
async function testServer(serverId) {
    try {
        await apiCall(`${API_BASE}/servers/${serverId}/test`, {
            method: 'POST'
        });
        showAlert('连接测试成功', 'success');
    } catch (error) {
        showAlert('连接测试失败', 'danger');
    }
}

// 同步设备
async function syncDevices(serverId) {
    try {
        await apiCall(`${API_BASE}/servers/${serverId}/sync-devices`, {
            method: 'POST'
        });
        showAlert('设备同步完成', 'success');
        loadDevices();
    } catch (error) {
        showAlert('设备同步失败: ' + error.message, 'danger');
    }
}

// 删除设备
function deleteDevice(deviceId) {
    if (!confirm('确定要删除这个设备吗？')) return;

    apiCall(`${API_BASE}/devices/${deviceId}`, { method: 'DELETE' })
        .then(() => {
            showAlert('设备删除成功', 'success');
            loadDevices();
        })
        .catch(error => {
            showAlert('删除设备失败', 'danger');
        });
}

// 删除服务器
function deleteServer(serverId) {
    if (!confirm('确定要删除这个服务器吗？')) return;

    apiCall(`${API_BASE}/servers/${serverId}`, { method: 'DELETE' })
        .then(() => {
            showAlert('服务器删除成功', 'success');
            loadServers();
        })
        .catch(error => {
            showAlert('删除服务器失败', 'danger');
        });
}

// 添加关联关系
async function addAssociation() {
    const deviceId = document.getElementById('select-device').value;
    const serverId = document.getElementById('select-server').value;
    const priority = document.getElementById('priority').value;

    if (!deviceId || !serverId) {
        showAlert('请选择设备和服务器', 'warning');
        return;
    }

    try {
        await apiCall(`${API_BASE}/devices/${deviceId}/servers/${serverId}`, {
            method: 'POST',
            body: JSON.stringify({ priority: parseInt(priority) })
        });

        showAlert('关联关系创建成功', 'success');

        // 重置表单
        document.getElementById('select-device').value = '';
        document.getElementById('select-server').value = '';
        document.getElementById('priority').value = '1';
    } catch (error) {
        showAlert('创建关联关系失败', 'danger');
    }
}

// 修改密码
function changePassword() {
    const oldPassword = document.getElementById('old-password').value;
    const newPassword = document.getElementById('new-password').value;
    const confirmPassword = document.getElementById('confirm-password').value;

    if (!oldPassword || !newPassword || !confirmPassword) {
        showAlert('请填写所有密码字段', 'warning');
        return;
    }

    if (newPassword !== confirmPassword) {
        showAlert('新密码确认不一致', 'warning');
        return;
    }

    apiCall(`${API_BASE}/auth/change-password`, {
        method: 'POST',
        body: JSON.stringify({
            old_password: oldPassword,
            new_password: newPassword
        })
    })
    .then(() => {
        showAlert('密码修改成功', 'success');
        // 清空表单
        document.getElementById('old-password').value = '';
        document.getElementById('new-password').value = '';
        document.getElementById('confirm-password').value = '';
    })
    .catch(error => {
        showAlert('密码修改失败', 'danger');
    });
}

// 显示登录模态框
function showLoginModal() {
    const modal = new bootstrap.Modal(document.getElementById('loginModal'));
    modal.show();
}