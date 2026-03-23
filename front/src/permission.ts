import router from './router'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import { useUserStore } from '@/store/user'

NProgress.configure({ showSpinner: false })

const whiteList = ['/login', '/home', '/', '/activity']

router.beforeEach((to, from, next) => {
    NProgress.start()

    const userStore = useUserStore()

    // check if path is in whitelist directly or starting with whitelist path
    const isWhiteListed = whiteList.some(path => to.path === path || to.path.startsWith('/activity/'))

    if (userStore.token) {
        if (to.path === '/login') {
            next({ path: '/' })
            NProgress.done()
        } else {
            next()
        }
    } else {
        if (isWhiteListed) {
            next()
        } else {
            next(`/login?redirect=${to.path}`)
            NProgress.done()
        }
    }
})

router.afterEach(() => {
    NProgress.done()
})
