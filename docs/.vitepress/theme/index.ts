import mediumZoom, { type Zoom } from 'medium-zoom';
import { defineComponent, h, nextTick, onMounted, onUnmounted, watch } from 'vue';
import DefaultTheme from 'vitepress/theme';
import { useRoute } from 'vitepress';
import './custom.css';

const Layout = defineComponent({
  setup() {
    const route = useRoute();
    let zoom: Zoom | undefined;

    const attachZoom = async () => {
      await nextTick();
      zoom?.detach();
      zoom?.attach('.home-screenshots img, .doc-screenshot');
    };

    onMounted(() => {
      zoom = mediumZoom({
        background: 'var(--vp-c-bg)',
        margin: 24,
      });
      void attachZoom();
    });

    watch(() => route.path, attachZoom, { flush: 'post' });

    onUnmounted(() => {
      zoom?.detach();
    });

    return () => h(DefaultTheme.Layout);
  },
});

export default {
  extends: DefaultTheme,
  Layout,
};
