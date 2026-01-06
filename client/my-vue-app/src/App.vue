<template>
  <div id="app">
    <h1>Metadata Viewer</h1>
    <MetadataList />
  </div>
</template>

<script>
import MetadataList from "./components/MetadataList.vue";

export default {
  components: { MetadataList },
};
</script>
