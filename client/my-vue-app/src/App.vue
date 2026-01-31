<template>
  <div id="app">
    <h1>Metadata Viewer</h1>
    <HelloWorld />
    <MetadataList />
  </div>
</template>

<script>
import MetadataList from "./components/MetadataList.vue";
import HelloWorld from "./components/HelloWorld.vue";

export default {
  components: { MetadataList, HelloWorld }
};
</script>
