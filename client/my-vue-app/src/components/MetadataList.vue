<template>
  <div>
    <h2>All Metadata</h2>
    <ul>
      <li v-for="item in metadata" :key="item.id">
        {{ item.fileName }} ({{ item.fileType }}) — Owner: {{ item.owner }}
      </li>
    </ul>
    <p v-if="metadata.length === 0">No metadata found.</p>
  </div>
</template>

<script>
import axios from "axios";

export default {
  data() {
    return {
      metadata: [],
    };
  },
  
  async mounted() {
    try {
      const metadataApi = axios.create({
        baseURL: "http://localhost:7676", // your Go server
      });

      // Fetch all keys
      const resIds = await metadataApi.get("/list_meta_ids");
      const allKeys = resIds.data;

      // Filter only objid keys and remove prefix
      const ids = allKeys
        .filter(key => key.startsWith("objid:"))
        .map(key => key.split(":")[1]);

      // Fetch metadata for each ID
      const metaList = await Promise.all(
        ids.map(async id => {
          try {
            const res = await metadataApi.get(`/read_meta?id=${encodeURIComponent(id)}`);
            return res.data;
          } catch (err) {
            console.warn(`Failed to fetch metadata for ID ${id}:`, err);
            return null;
          }
        })
      );

      // Only keep valid metadata
      this.metadata = metaList.filter(item => item !== null);
    } catch (err) {
      console.error("Failed to fetch metadata:", err);
    }
  },
};
</script>
