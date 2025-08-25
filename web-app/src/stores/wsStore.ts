import { defineStore } from "pinia";
import { Centrifuge, Subscription } from "centrifuge";
import { useGameStore } from "./gameStore";
import type { GameUpdate } from "../models/common";

interface CentrifugeStoreState {
  client: Centrifuge | null;
  sub: Subscription | null;
  connected: boolean;
  subscribed: boolean;
}

export const useWSStore = defineStore("centrifuge", {
  state: (): CentrifugeStoreState => ({
    client: null,
    sub: null,
    connected: false,
    subscribed: false,
  }),
  actions: {
    connect() {
      const cf = new Centrifuge(
        `${import.meta.env.VITE_WS_BASE_URL}/api/v1/games/updates`,
        {
          token: "test",
        }
      );

      cf.on("connected", () => {
        this.connected = true;
      });
      cf.on("disconnected", () => {
        this.connected = false;
      });
      cf.on("error", (ctx) => {
        console.log(ctx)
      })
      cf.connect();

      this.client = cf;
    },
    disconnect() {
      this.sub?.unsubscribe();
      this.sub = null;
      this.client?.disconnect();
      this.client = null;
    },
    subscribe(gameInstanceId: number) {
      if (this.client == null) {
        return;
      }

      this.subscribed = false;

      const channelName = `game_${gameInstanceId}`;
      let sub = this.client.getSubscription(channelName);
      if (!sub) {
        sub = this.client.newSubscription(channelName, {
          data: {
            user_id: 1,
            game_id: gameInstanceId,
          },
        });
      }

      sub.on("subscribed", (ctx) => {
        this.subscribed = true;
        let upd = ctx.data as GameUpdate;
        const gameStore = useGameStore();
        gameStore.setGameInstance(upd);
      });
      sub.on("unsubscribed", (ctx) => {
        this.subscribed = false;
      });
      sub.on("publication", (ctx) => {
        const gameStore = useGameStore();
        let data = ctx.data as GameUpdate;
        gameStore.updateGameStatus(data);
      });
      sub.on("error", (ctx) => {
        console.log(ctx);
      });

      sub.subscribe();
      this.sub = sub;
    },
    unsubsribe() {
      if (this.sub == null) {
        return;
      }

      this.sub.unsubscribe();
      this.sub = null;
    }
  },
});
