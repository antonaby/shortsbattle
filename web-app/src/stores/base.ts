import {  defineStore } from 'pinia'

export const useBaseStore = defineStore('base', {
  state: () => ({
    count: 0,
  }),
  actions: {
    increment() {
      this.count++
    }
  }
})
