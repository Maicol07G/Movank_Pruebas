<script>
let online = navigator.onLine;
let products = [];
let message = '';
async function load() {
  try {
    const r = await fetch('http://localhost:8080/v1/products');
    products = await r.json();
    message = 'Catálogo cargado';
  } catch { message = 'Sin conexión: usar catálogo cacheado/IndexedDB'; }
}
load();
</script>

<svelte:head><title>MOVANK</title></svelte:head>
<main>
  <h1>MOVANK</h1>
  <p>Estado: {online ? '🟢 Online' : '🔴 Offline'}</p>
  <p>{message}</p>
  <h2>Catálogo</h2>
  {#if products.length}
    {#each products as p}
      <article><strong>{p.name}</strong><span>${p.price_cents}</span></article>
    {/each}
  {:else}
    <p>No hay productos cargados todavía.</p>
  {/if}
</main>

<style>
main{max-width:900px;margin:40px auto;padding:20px;font-family:system-ui}
article{display:flex;justify-content:space-between;padding:16px;margin:10px 0;border:1px solid #ddd;border-radius:10px}
</style>
