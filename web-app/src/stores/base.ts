import {  defineStore } from 'pinia'
import axios from 'axios';
import { Centrifuge } from 'centrifuge';

interface BaseState {
  loading: boolean
  msg: string,
  centrifuge: Centrifuge | null,
  connected: boolean,
  messages: any[]
}

export const useBaseStore = defineStore('base', {
  state: (): BaseState => ({
    loading: false,
    msg: '',
    centrifuge: null,
    connected: false,
    messages: []
  }),
  actions: {
    initCentriguge() {
      if (this.centrifuge) return; // avoid double init

      const centrifuge = new Centrifuge('ws://localhost:8080/v1/enter');

      // Allocate Subscription to a channel.
      const sub = centrifuge.newSubscription('news');

      // React on `news` channel real-time publications.
      sub.on('publication', function(ctx) {
          console.log(ctx.data);
      });

      // Trigger subscribe process.
      sub.subscribe();

      // Trigger actual connection establishement.
      centrifuge.connect();
      this.centrifuge = centrifuge;
    },

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
