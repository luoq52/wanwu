import Vue from 'vue'
import Vuex from 'vuex'
import VuexPersistence from 'vuex-persist'
import { login } from './module/login'
import { user } from './module/user'
import { app } from './module/app'
import { workflow } from './module/workflow'


Vue.use(Vuex)
// 用户信息持久化
const vuexLocal = new VuexPersistence({
    key:'access_cert',
    storage: window.localStorage,
    modules: ['user']
})
//知识库全选权限持久化
const permissionLocal = new VuexPersistence({
    key:'permission_data',
    storage: window.localStorage,
    modules: ['app'],
    filter:(mutation) => {
        return mutation.type === 'SET_PERMISSION_TYPE' || 
               mutation.type === 'CLEAR_PERMISSION_TYPE'
    },
    restoreState:(key,storage) => {
        const userData = localStorage.getItem('access_cert')
        if (!userData) {
            return {}
        }
        return JSON.parse(storage.getItem(key) || '{}')
    }
})

export const store = new Vuex.Store({
    modules: {
        login,
        user,
        app,
        workflow,
    },
    plugins: [vuexLocal.plugin, permissionLocal.plugin]
})
