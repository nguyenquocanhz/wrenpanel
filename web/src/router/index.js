import { createRouter, createWebHistory } from 'vue-router'

import VhostsView from '../views/Vhosts/VhostsView.vue'
import PhpView from '../views/PhpManager/PhpView.vue'
import AppsView from '../views/Apps/AppsView.vue'
import NodeView from '../views/NodeManager/NodeView.vue'
import PythonView from '../views/PythonManager/PythonView.vue'
import DatabasesView from '../views/Databases/DatabasesView.vue'
import FTPView from '../views/FTP/FTPView.vue'
import SSLView from '../views/SSL/SSLView.vue'
import BackupsView from '../views/Backups/BackupsView.vue'
import AccountsView from '../views/Accounts/AccountsView.vue'
import SystemView from '../views/System/SystemView.vue'
import LoginView from '../views/Auth/LoginView.vue'

const routes = [
  { path: '/login', name: 'login', component: LoginView },
  { path: '/', redirect: '/vhosts' },
  { path: '/vhosts', name: 'vhosts', component: VhostsView },
  { path: '/apps', name: 'apps', component: AppsView },
  { path: '/php', name: 'php', component: PhpView },
  { path: '/node', name: 'node', component: NodeView },
  { path: '/python', name: 'python', component: PythonView },
  { path: '/databases', name: 'databases', component: DatabasesView },
  { path: '/ftp', name: 'ftp', component: FTPView },
  { path: '/ssl', name: 'ssl', component: SSLView },
  { path: '/backups', name: 'backups', component: BackupsView },
  { path: '/accounts', name: 'accounts', component: AccountsView },
  { path: '/system', name: 'system', component: SystemView },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})
