<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute } from 'vue-router'

import {
  changePassword,
  createAddress,
  createTicketBuyer,
  deleteAddress,
  deleteTicketBuyer,
  getProfile,
  listAddresses,
  listTicketBuyers,
  setDefaultAddress,
  setDefaultTicketBuyer,
  updateAddress,
  updateProfile,
  updateTicketBuyer,
  type AddressPayload,
  type TicketBuyerPayload,
  type UserProfilePayload,
} from '@/api/user'
import { useUserStore } from '@/store/user'

type ProfileTab = 'info' | 'attendees' | 'addresses'
type DialogMode = 'create' | 'edit'

const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const phoneRegex = /^1\d{10}$/
const idCardRegex = /(^\d{15}$)|(^\d{17}[\dXx]$)/

const userStore = useUserStore()
const route = useRoute()

const activeTab = ref<ProfileTab>('info')
const pageLoading = ref(false)
const profileSubmitting = ref(false)
const attendeesLoading = ref(false)
const attendeeSubmitting = ref(false)
const addressesLoading = ref(false)
const addressSubmitting = ref(false)
const passwordDialogVisible = ref(false)
const passwordSubmitting = ref(false)
const attendeeDialogVisible = ref(false)
const addressDialogVisible = ref(false)
const attendeeMode = ref<DialogMode>('create')
const addressMode = ref<DialogMode>('create')

const attendees = ref<TicketBuyerPayload[]>([])
const addresses = ref<AddressPayload[]>([])

const profileForm = reactive({
  nickname: '',
  avatar: '',
  email: '',
})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const attendeeForm = reactive({
  buyerId: '',
  name: '',
  idCard: '',
  phone: '',
  isDefault: false,
})

const addressForm = reactive({
  addressId: '',
  receiverName: '',
  receiverPhone: '',
  province: '',
  city: '',
  district: '',
  detail: '',
  isDefault: false,
})

const avatarPreview = computed(() => profileForm.avatar || userStore.userInfo.avatar || defaultAvatar)
const attendeeDialogTitle = computed(() => attendeeMode.value === 'create' ? '新增观演人' : '编辑观演人')
const addressDialogTitle = computed(() => addressMode.value === 'create' ? '新增地址' : '编辑地址')

function resolveProfileTab(value: unknown): ProfileTab {
  if (value === 'attendees' || value === 'addresses' || value === 'info') {
    return value
  }

  return 'info'
}

function handleTabSelect(value: string) {
  activeTab.value = value as ProfileTab
}

function syncProfileForm(profile: UserProfilePayload | null = null) {
  const source = profile ?? {
    id: userStore.userInfo.id,
    phone: userStore.userInfo.phone,
    email: userStore.userInfo.email,
    nickname: userStore.userInfo.nickname,
    avatar: userStore.userInfo.avatar,
    status: userStore.userInfo.status,
    isVerified: userStore.userInfo.isVerified,
    realName: userStore.userInfo.realName,
    createdAt: userStore.userInfo.createdAt,
  }

  profileForm.nickname = source.nickname || ''
  profileForm.avatar = source.avatar || ''
  profileForm.email = source.email || ''
}

function applyProfile(profile: UserProfilePayload) {
  userStore.setUserInfo({
    id: profile.id,
    phone: profile.phone,
    email: profile.email,
    nickname: profile.nickname,
    avatar: profile.avatar,
    status: profile.status,
    isVerified: profile.isVerified,
    isRealName: profile.isVerified,
    realName: profile.realName || '',
    createdAt: profile.createdAt || '',
  })
  syncProfileForm(profile)
}

function formatRegion(address: AddressPayload) {
  return [address.province, address.city, address.district].filter(Boolean).join(' / ')
}

function formatFullAddress(address: AddressPayload) {
  return [address.province, address.city, address.district, address.detail].filter(Boolean).join(' ')
}

function resetPasswordForm() {
  passwordForm.oldPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
}

function resetAttendeeForm() {
  attendeeForm.buyerId = ''
  attendeeForm.name = ''
  attendeeForm.idCard = ''
  attendeeForm.phone = ''
  attendeeForm.isDefault = false
}

function resetAddressForm() {
  addressForm.addressId = ''
  addressForm.receiverName = ''
  addressForm.receiverPhone = ''
  addressForm.province = ''
  addressForm.city = ''
  addressForm.district = ''
  addressForm.detail = ''
  addressForm.isDefault = false
}

function validateProfileForm() {
  if (!profileForm.nickname.trim()) {
    ElMessage.warning('请输入昵称')
    return false
  }

  if (profileForm.email.trim() && !emailRegex.test(profileForm.email.trim())) {
    ElMessage.warning('邮箱格式不正确')
    return false
  }

  return true
}

function validatePasswordForm() {
  if (!passwordForm.oldPassword) {
    ElMessage.warning('请输入当前密码')
    return false
  }

  if (!passwordForm.newPassword || passwordForm.newPassword.length < 6) {
    ElMessage.warning('新密码长度不能少于 6 位')
    return false
  }

  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    ElMessage.warning('两次输入的新密码不一致')
    return false
  }

  return true
}

function validateAttendeeForm() {
  if (!attendeeForm.name.trim()) {
    ElMessage.warning('请输入观演人姓名')
    return false
  }

  if (attendeeForm.phone.trim() && !phoneRegex.test(attendeeForm.phone.trim())) {
    ElMessage.warning('手机号格式不正确')
    return false
  }

  if (attendeeForm.idCard.trim() && !idCardRegex.test(attendeeForm.idCard.trim())) {
    ElMessage.warning('身份证号格式不正确')
    return false
  }

  return true
}

function validateAddressForm() {
  if (!addressForm.receiverName.trim()) {
    ElMessage.warning('请输入收货人姓名')
    return false
  }

  if (!phoneRegex.test(addressForm.receiverPhone.trim())) {
    ElMessage.warning('请输入正确的收货人手机号')
    return false
  }

  if (!addressForm.province.trim() || !addressForm.city.trim() || !addressForm.district.trim()) {
    ElMessage.warning('请填写完整的省市区信息')
    return false
  }

  if (!addressForm.detail.trim()) {
    ElMessage.warning('请输入详细地址')
    return false
  }

  return true
}

async function loadProfile(showLoading = true) {
  if (showLoading) {
    pageLoading.value = true
  }

  try {
    const profile = await getProfile()
    applyProfile(profile)
  } finally {
    if (showLoading) {
      pageLoading.value = false
    }
  }
}

async function loadAttendees(showLoading = true) {
  if (showLoading) {
    attendeesLoading.value = true
  }

  try {
    const resp = await listTicketBuyers()
    attendees.value = resp.ticketBuyers
  } finally {
    if (showLoading) {
      attendeesLoading.value = false
    }
  }
}

async function loadAddressList(showLoading = true) {
  if (showLoading) {
    addressesLoading.value = true
  }

  try {
    const resp = await listAddresses()
    addresses.value = resp.addresses
  } finally {
    if (showLoading) {
      addressesLoading.value = false
    }
  }
}

async function initializePage() {
  pageLoading.value = true
  attendeesLoading.value = true
  addressesLoading.value = true

  try {
    const [profile, buyerResp, addressResp] = await Promise.all([
      getProfile(),
      listTicketBuyers(),
      listAddresses(),
    ])

    applyProfile(profile)
    attendees.value = buyerResp.ticketBuyers
    addresses.value = addressResp.addresses
  } finally {
    pageLoading.value = false
    attendeesLoading.value = false
    addressesLoading.value = false
  }
}

async function submitProfile() {
  if (!validateProfileForm()) {
    return
  }

  profileSubmitting.value = true
  try {
    const profile = await updateProfile({
      nickname: profileForm.nickname.trim(),
      avatar: profileForm.avatar.trim(),
      email: profileForm.email.trim(),
    })
    applyProfile(profile)
    ElMessage.success('个人信息已更新')
  } finally {
    profileSubmitting.value = false
  }
}

function openPasswordDialog() {
  resetPasswordForm()
  passwordDialogVisible.value = true
}

async function submitPassword() {
  if (!validatePasswordForm()) {
    return
  }

  passwordSubmitting.value = true
  try {
    await changePassword({
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword,
    })
    passwordDialogVisible.value = false
    resetPasswordForm()
    ElMessage.success('密码已修改')
  } finally {
    passwordSubmitting.value = false
  }
}

function openCreateAttendeeDialog() {
  attendeeMode.value = 'create'
  resetAttendeeForm()
  attendeeDialogVisible.value = true
}

function openEditAttendeeDialog(attendee: TicketBuyerPayload) {
  attendeeMode.value = 'edit'
  attendeeForm.buyerId = attendee.id
  attendeeForm.name = attendee.name
  attendeeForm.idCard = ''
  attendeeForm.phone = attendee.phone || ''
  attendeeForm.isDefault = attendee.isDefault
  attendeeDialogVisible.value = true
}

async function submitAttendee() {
  if (!validateAttendeeForm()) {
    return
  }

  attendeeSubmitting.value = true
  try {
    const payload = {
      name: attendeeForm.name.trim(),
      idCard: attendeeForm.idCard.trim(),
      phone: attendeeForm.phone.trim(),
      isDefault: attendeeForm.isDefault,
    }

    if (attendeeMode.value === 'create') {
      await createTicketBuyer(payload)
      ElMessage.success('观演人已新增')
    } else {
      await updateTicketBuyer({
        buyerId: attendeeForm.buyerId,
        ...payload,
      })
      ElMessage.success('观演人已更新')
    }

    attendeeDialogVisible.value = false
    await loadAttendees()
  } finally {
    attendeeSubmitting.value = false
  }
}

async function handleDeleteAttendee(attendee: TicketBuyerPayload) {
  try {
    await ElMessageBox.confirm(`确认删除观演人“${attendee.name}”吗？`, '删除观演人', {
      type: 'warning',
    })
  } catch {
    return
  }

  await deleteTicketBuyer(attendee.id)
  ElMessage.success('观演人已删除')
  await loadAttendees()
}

async function handleSetDefaultAttendee(attendee: TicketBuyerPayload) {
  if (attendee.isDefault) {
    return
  }

  await setDefaultTicketBuyer(attendee.id)
  ElMessage.success('已设为默认观演人')
  await loadAttendees()
}

function openCreateAddressDialog() {
  addressMode.value = 'create'
  resetAddressForm()
  addressDialogVisible.value = true
}

function openEditAddressDialog(address: AddressPayload) {
  addressMode.value = 'edit'
  addressForm.addressId = address.id
  addressForm.receiverName = address.receiverName
  addressForm.receiverPhone = address.receiverPhone
  addressForm.province = address.province
  addressForm.city = address.city
  addressForm.district = address.district
  addressForm.detail = address.detail
  addressForm.isDefault = address.isDefault
  addressDialogVisible.value = true
}

async function submitAddress() {
  if (!validateAddressForm()) {
    return
  }

  addressSubmitting.value = true
  try {
    const payload = {
      receiverName: addressForm.receiverName.trim(),
      receiverPhone: addressForm.receiverPhone.trim(),
      province: addressForm.province.trim(),
      city: addressForm.city.trim(),
      district: addressForm.district.trim(),
      detail: addressForm.detail.trim(),
      isDefault: addressForm.isDefault,
    }

    if (addressMode.value === 'create') {
      await createAddress(payload)
      ElMessage.success('地址已新增')
    } else {
      await updateAddress({
        addressId: addressForm.addressId,
        ...payload,
      })
      ElMessage.success('地址已更新')
    }

    addressDialogVisible.value = false
    await loadAddressList()
  } finally {
    addressSubmitting.value = false
  }
}

async function handleDeleteAddress(address: AddressPayload) {
  try {
    await ElMessageBox.confirm(`确认删除地址“${formatFullAddress(address)}”吗？`, '删除地址', {
      type: 'warning',
    })
  } catch {
    return
  }

  await deleteAddress(address.id)
  ElMessage.success('地址已删除')
  await loadAddressList()
}

async function refreshProfile() {
  await loadProfile()
}

async function handleSetDefaultAddress(address: AddressPayload) {
  if (address.isDefault) {
    return
  }

  await setDefaultAddress(address.id)
  ElMessage.success('已设为默认地址')
  await loadAddressList()
}

onMounted(() => {
  syncProfileForm()
  void initializePage()
})

watch(
  () => route.query.tab,
  (value) => {
    const tabValue = Array.isArray(value) ? value[0] : value
    activeTab.value = resolveProfileTab(tabValue)
  },
  { immediate: true },
)
</script>

<template>
  <div class="profile-container wrapper">
    <aside class="sidebar">
      <el-menu :default-active="activeTab" class="side-menu" @select="handleTabSelect">
        <el-menu-item index="info">
          <el-icon><User /></el-icon>
          <span>个人信息</span>
        </el-menu-item>
        <el-menu-item index="attendees">
          <el-icon><Avatar /></el-icon>
          <span>常用观演人</span>
        </el-menu-item>
        <el-menu-item index="addresses">
          <el-icon><Location /></el-icon>
          <span>收货地址</span>
        </el-menu-item>
      </el-menu>
    </aside>

    <main class="main-content">
      <el-card v-loading="pageLoading" shadow="never" class="profile-card">
        <section v-if="activeTab === 'info'" class="panel-section">
          <div class="tab-header">
            <h3 class="tab-title">个人信息</h3>
            <el-button plain @click="openPasswordDialog">
              <el-icon><Lock /></el-icon>
              <span>修改密码</span>
            </el-button>
          </div>

          <el-form label-width="108px" class="profile-form">
            <el-form-item label="头像">
              <div class="avatar-block">
                <el-avatar :size="68" :src="avatarPreview" />
                <div class="avatar-fields">
                  <el-input
                    v-model="profileForm.avatar"
                    placeholder="输入头像图片地址"
                    clearable
                  />
                  <div class="helper-text">当前后端仅保存头像链接，这里直接填写可访问的图片 URL。</div>
                </div>
              </div>
            </el-form-item>

            <el-form-item label="昵称" required>
              <el-input v-model="profileForm.nickname" maxlength="30" placeholder="请输入昵称" />
            </el-form-item>

            <el-form-item label="邮箱">
              <el-input v-model="profileForm.email" placeholder="请输入邮箱" clearable />
            </el-form-item>

            <el-form-item label="手机号">
              <div class="readonly-text">{{ userStore.userInfo.phone || '-' }}</div>
            </el-form-item>

            <el-form-item label="实名状态">
              <div class="verify-block">
                <el-tag :type="userStore.userInfo.isVerified ? 'success' : 'info'">
                  {{ userStore.userInfo.isVerified ? '已实名' : '未实名' }}
                </el-tag>
                <span class="verify-text">
                  {{
                    userStore.userInfo.isVerified
                      ? `实名用户：${userStore.userInfo.realName || '已脱敏'}`
                      : '当前页面暂未提供实名提交通道'
                  }}
                </span>
              </div>
            </el-form-item>

            <el-form-item label="注册时间">
              <div class="readonly-text">{{ userStore.userInfo.createdAt || '-' }}</div>
            </el-form-item>

            <el-form-item>
              <el-button type="primary" :loading="profileSubmitting" @click="submitProfile">保存修改</el-button>
              <el-button @click="refreshProfile" :disabled="profileSubmitting">
                <el-icon><RefreshRight /></el-icon>
                <span>刷新资料</span>
              </el-button>
            </el-form-item>
          </el-form>
        </section>

        <section v-else-if="activeTab === 'attendees'" class="panel-section">
          <div class="tab-header">
            <h3 class="tab-title">常用观演人管理</h3>
            <el-button type="primary" @click="openCreateAttendeeDialog">
              <el-icon><Plus /></el-icon>
              <span>新增观演人</span>
            </el-button>
          </div>

          <el-table v-loading="attendeesLoading" :data="attendees" border style="width: 100%">
            <el-table-column prop="name" label="姓名" min-width="140" />
            <el-table-column prop="phone" label="手机号" min-width="150">
              <template #default="{ row }">
                {{ row.phone || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="idCard" label="身份证号" min-width="200">
              <template #default="{ row }">
                {{ row.idCard || '-' }}
              </template>
            </el-table-column>
            <el-table-column label="默认" width="90" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.isDefault" type="success" size="small">默认</el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="创建时间" min-width="170" />
            <el-table-column label="操作" min-width="180" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openEditAttendeeDialog(row)">编辑</el-button>
                <el-button link type="primary" :disabled="row.isDefault" @click="handleSetDefaultAttendee(row)">
                  设为默认
                </el-button>
                <el-button link type="danger" @click="handleDeleteAttendee(row)">删除</el-button>
              </template>
            </el-table-column>

            <template #empty>
              <el-empty description="暂无观演人，新增后可在下单页直接选择" />
            </template>
          </el-table>
        </section>

        <section v-else class="panel-section">
          <div class="tab-header">
            <h3 class="tab-title">收货地址管理</h3>
            <el-button type="primary" @click="openCreateAddressDialog">
              <el-icon><Plus /></el-icon>
              <span>新增地址</span>
            </el-button>
          </div>

          <el-table v-loading="addressesLoading" :data="addresses" border style="width: 100%">
            <el-table-column prop="receiverName" label="收货人" min-width="120" />
            <el-table-column prop="receiverPhone" label="手机号" min-width="140" />
            <el-table-column label="省市区" min-width="180">
              <template #default="{ row }">
                {{ formatRegion(row) }}
              </template>
            </el-table-column>
            <el-table-column prop="detail" label="详细地址" min-width="260" />
            <el-table-column label="默认" width="90" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.isDefault" type="success" size="small">默认</el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="创建时间" min-width="170" />
            <el-table-column label="操作" min-width="180" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openEditAddressDialog(row)">编辑</el-button>
                <el-button link type="primary" :disabled="row.isDefault" @click="handleSetDefaultAddress(row)">
                  设为默认
                </el-button>
                <el-button link type="danger" @click="handleDeleteAddress(row)">删除</el-button>
              </template>
            </el-table-column>

            <template #empty>
              <el-empty description="暂无收货地址，纸质票场景下可直接复用" />
            </template>
          </el-table>
        </section>
      </el-card>
    </main>

    <el-dialog
      v-model="passwordDialogVisible"
      title="修改密码"
      width="420px"
      destroy-on-close
    >
      <el-form label-width="96px">
        <el-form-item label="当前密码" required>
          <el-input v-model="passwordForm.oldPassword" type="password" show-password />
        </el-form-item>
        <el-form-item label="新密码" required>
          <el-input v-model="passwordForm.newPassword" type="password" show-password />
        </el-form-item>
        <el-form-item label="确认密码" required>
          <el-input v-model="passwordForm.confirmPassword" type="password" show-password />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="passwordSubmitting" @click="submitPassword">确认修改</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="attendeeDialogVisible"
      :title="attendeeDialogTitle"
      width="460px"
      destroy-on-close
    >
      <el-form label-width="96px">
        <el-form-item label="姓名" required>
          <el-input v-model="attendeeForm.name" maxlength="20" placeholder="请输入观演人姓名" />
        </el-form-item>
        <el-form-item label="身份证号">
          <el-input
            v-model="attendeeForm.idCard"
            maxlength="18"
            :placeholder="attendeeMode === 'create' ? '请输入身份证号' : '如不修改可留空'"
          />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="attendeeForm.phone" maxlength="11" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="默认观演人">
          <el-switch v-model="attendeeForm.isDefault" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="attendeeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="attendeeSubmitting" @click="submitAttendee">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="addressDialogVisible"
      :title="addressDialogTitle"
      width="520px"
      destroy-on-close
    >
      <el-form label-width="96px">
        <el-form-item label="收货人" required>
          <el-input v-model="addressForm.receiverName" maxlength="20" placeholder="请输入收货人姓名" />
        </el-form-item>
        <el-form-item label="手机号" required>
          <el-input v-model="addressForm.receiverPhone" maxlength="11" placeholder="请输入收货手机号" />
        </el-form-item>
        <el-form-item label="省份" required>
          <el-input v-model="addressForm.province" placeholder="如：北京市" />
        </el-form-item>
        <el-form-item label="城市" required>
          <el-input v-model="addressForm.city" placeholder="如：北京市" />
        </el-form-item>
        <el-form-item label="区县" required>
          <el-input v-model="addressForm.district" placeholder="如：朝阳区" />
        </el-form-item>
        <el-form-item label="详细地址" required>
          <el-input
            v-model="addressForm.detail"
            type="textarea"
            :rows="3"
            placeholder="请输入详细地址"
          />
        </el-form-item>
        <el-form-item label="默认地址">
          <el-switch v-model="addressForm.isDefault" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="addressDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="addressSubmitting" @click="submitAddress">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.wrapper {
  max-width: 1200px;
  margin: 0 auto;
  padding: 30px 0;
  display: flex;
  gap: 20px;
}

.sidebar {
  width: 220px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 10px 30px rgba(15, 23, 42, 0.05);
  overflow: hidden;
  height: max-content;
}

.side-menu {
  border-right: none;
}

.main-content {
  flex: 1;
  min-width: 0;
}

.profile-card {
  border-radius: 10px;
  min-height: 560px;
}

.panel-section {
  min-height: 420px;
}

.tab-title {
  margin: 0;
  font-size: 18px;
  color: #1f2937;
  border-left: 4px solid #409eff;
  padding-left: 10px;
}

.tab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  gap: 12px;
}

.profile-form {
  max-width: 680px;
}

.avatar-block {
  display: flex;
  align-items: center;
  gap: 18px;
  width: 100%;
}

.avatar-fields {
  flex: 1;
}

.helper-text {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}

.readonly-text {
  color: #303133;
  line-height: 32px;
}

.verify-block {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.verify-text {
  color: #606266;
}

@media (max-width: 960px) {
  .wrapper {
    padding: 20px 16px 32px;
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
  }

  .tab-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .avatar-block {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
