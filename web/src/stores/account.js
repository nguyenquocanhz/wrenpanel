import { defineStore } from 'pinia'
import { accountsApi } from '../api/accounts'

export const useAccountStore = defineStore('account', {
  state: () => ({
    accounts: [],
    selectedAccountId: null,
    loading: false,
    error: null,
  }),
  getters: {
    currentAccount: (state) => {
      if (!state.selectedAccountId) return state.accounts[0] || null
      return state.accounts.find((a) => a.id === state.selectedAccountId) || state.accounts[0] || null
    },
  },
  actions: {
    async fetchAccounts() {
      this.loading = true
      try {
        const list = await accountsApi.list()
        this.accounts = list || []
        if (this.accounts.length > 0 && !this.selectedAccountId) {
          this.selectedAccountId = this.accounts[0].id
        }
      } catch (err) {
        this.error = err.message
      } finally {
        this.loading = false
      }
    },
    selectAccount(id) {
      this.selectedAccountId = id
    },
  },
})
