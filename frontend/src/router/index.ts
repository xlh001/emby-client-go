import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { userStore } from '@/stores/user'

// 页面组件
const Dashboard = () => import('@/views/Dashboard.vue')
const Login = () => import('@/views/Login.vue')
const ServerList = () => import('@/views/servers/ServerList.vue')
const ServerForm = () => import('@/views/servers/ServerForm.vue')
const ServerDetail = () => import('@/views/servers/ServerDetail.vue')
const LibraryBrowser = () => import('@/views/media/LibraryBrowser.vue')
const MediaSearch = () => import('@/views/media/MediaSearch.vue')
const MediaPlayer = () => import('@/views/media/MediaPlayer.vue')
const UserList = () => import('@/views/users/UserList.vue')
const UserForm = () => import('@/views/users/UserForm.vue')
const Profile = () => import('@/views/settings/Profile.vue')
const System = () => import('@/views/settings/System.vue')

// 路由配置
const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: {
      title: '登录',
      requiresAuth: false,
    },
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard,
    meta: {
      title: '仪表盘',
      icon: 'dashboard',
      requiresAuth: true,
    },
  },
  {
    path: '/servers',
    name: 'ServerList',
    component: ServerList,
    meta: {
      title: '服务器管理',
      icon: 'server',
      requiresAuth: true,
    },
  },
  {
    path: '/servers/add',
    name: 'ServerAdd',
    component: ServerForm,
    meta: {
      title: '添加服务器',
      requiresAuth: true,
    },
  },
  {
    path: '/servers/:id',
    name: 'ServerDetail',
    component: ServerDetail,
    meta: {
      title: '服务器详情',
      requiresAuth: true,
    },
  },
  {
    path: '/media/libraries',
    name: 'LibraryBrowser',
    component: LibraryBrowser,
    meta: {
      title: '媒体库',
      icon: 'folder',
      requiresAuth: true,
    },
  },
  {
    path: '/media/search',
    name: 'MediaSearch',
    component: MediaSearch,
    meta: {
      title: '媒体搜索',
      icon: 'search',
      requiresAuth: true,
    },
  },
  {
    path: '/media/player',
    name: 'MediaPlayer',
    component: MediaPlayer,
    meta: {
      title: '播放器',
      requiresAuth: true,
    },
  },
  {
    path: '/users',
    name: 'UserList',
    component: UserList,
    meta: {
      title: '用户管理',
      icon: 'user',
      requiresAuth: true,
      roles: ['admin'],
    },
  },
  {
    path: '/users/add',
    name: 'UserAdd',
    component: UserForm,
    meta: {
      title: '添加用户',
      requiresAuth: true,
      roles: ['admin'],
    },
  },
  {
    path: '/users/:id',
    name: 'UserEdit',
    component: UserForm,
    meta: {
      title: '编辑用户',
      requiresAuth: true,
      roles: ['admin'],
    },
  },
  {
    path: '/profile',
    name: 'Profile',
    component: Profile,
    meta: {
      title: '个人资料',
      requiresAuth: true,
    },
  },
  {
    path: '/settings',
    name: 'System',
    component: System,
    meta: {
      title: '系统设置',
      icon: 'setting',
      requiresAuth: true,
      roles: ['admin'],
    },
  },
]

// 创建路由实例
const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const store = userStore()

  // 设置页面标题
  document.title = to.meta.title ? `${to.meta.title} - Emby Manager` : 'Emby Manager'

  // 检查是否需要认证
  if (to.meta.requiresAuth && !store.isLoggedIn) {
    next('/login')
    return
  }

  // 检查角色权限
  if (to.meta.roles && !to.meta.roles.includes(store.user?.role || '')) {
    next('/dashboard')
    return
  }

  next()
})

export default router