import {  defineStore } from 'pinia'
import axios from 'axios';

export const useBaseStore = defineStore('base', {
  state: () => ({
    loading: false,
    msg: '',
  }),
  actions: {
    async getMessage() {
      this.loading = true;

      try {
        const response = await axios.get('http://localhost:8080/');
        this.msg = response.data.message;
      } catch (error) {
        if (error instanceof Error) {
          this.msg = error.message;
        } else {
          this.msg = 'Failed to fetch message';
        }
      } finally {
        this.loading = false;
      }
    }
  }
});
