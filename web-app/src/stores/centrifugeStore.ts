import { defineStore } from "pinia";
import { Centrifuge, Subscription } from "centrifuge";
import { useGameStore } from "./gameStore";

interface CentrifugeStoreState {
  client: Centrifuge | null;
  sub: Subscription | null;
  connected: boolean;
  subscribed: boolean;
}

export const useCentrifugeStore = defineStore("centrifuge", {
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

      sub.on("subscribed", () => {
        this.subscribed = true;
      });
      sub.on("unsubscribed", () => {
        this.subscribed = false;
      });

      sub.on("publication", (ctx) => {
        const gameStore = useGameStore();
        gameStore.updateGameStatus(ctx.data.status);
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
    },
    async publish(data: any) {
      if (this.sub == null) {
        return;
      }

      await this.sub.publish(data);
    },
  },
});
