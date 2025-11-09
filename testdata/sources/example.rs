use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};
use tokio::time::sleep;

/// Represents the result of a cache operation
#[derive(Debug, Clone)]
pub enum CacheResult<T> {
    Hit(T),
    Miss,
    Expired,
}

/// Entry in the cache with expiration tracking
#[derive(Debug, Clone)]
struct CacheEntry<T> {
    value: T,
    inserted_at: Instant,
    ttl: Duration,
}

impl<T: Clone> CacheEntry<T> {
    fn new(value: T, ttl: Duration) -> Self {
        Self {
            value,
            inserted_at: Instant::now(),
            ttl,
        }
    }

    fn is_expired(&self) -> bool {
        self.inserted_at.elapsed() > self.ttl
    }
}

/// Thread-safe in-memory cache with TTL support
pub struct Cache<K, V>
where
    K: Eq + std::hash::Hash + Clone,
    V: Clone,
{
    store: Arc<Mutex<HashMap<K, CacheEntry<V>>>>,
    default_ttl: Duration,
    max_size: usize,
    stats: Arc<Mutex<CacheStats>>,
}

/// Statistics for cache operations
#[derive(Debug, Default, Clone)]
pub struct CacheStats {
    hits: u64,
    misses: u64,
    evictions: u64,
    expirations: u64,
}

impl<K, V> Cache<K, V>
where
    K: Eq + std::hash::Hash + Clone,
    V: Clone,
{
    /// Create a new cache with default TTL and maximum size
    pub fn new(default_ttl: Duration, max_size: usize) -> Self {
        Self {
            store: Arc::new(Mutex::new(HashMap::new())),
            default_ttl,
            max_size,
            stats: Arc::new(Mutex::new(CacheStats::default())),
        }
    }

    /// Insert a value into the cache with default TTL
    pub fn insert(&self, key: K, value: V) {
        self.insert_with_ttl(key, value, self.default_ttl);
    }

    /// Insert a value with custom TTL
    pub fn insert_with_ttl(&self, key: K, value: V, ttl: Duration) {
        let mut store = self.store.lock().unwrap();
        
        // Evict oldest entry if at capacity
        if store.len() >= self.max_size && !store.contains_key(&key) {
            if let Some(oldest_key) = store.keys().next().cloned() {
                store.remove(&oldest_key);
                let mut stats = self.stats.lock().unwrap();
                stats.evictions += 1;
            }
        }

        store.insert(key, CacheEntry::new(value, ttl));
    }

    /// Get a value from the cache
    pub fn get(&self, key: &K) -> CacheResult<V> {
        let mut store = self.store.lock().unwrap();
        let mut stats = self.stats.lock().unwrap();

        match store.get(key) {
            Some(entry) => {
                if entry.is_expired() {
                    store.remove(key);
                    stats.expirations += 1;
                    stats.misses += 1;
                    CacheResult::Expired
                } else {
                    stats.hits += 1;
                    CacheResult::Hit(entry.value.clone())
                }
            }
            None => {
                stats.misses += 1;
                CacheResult::Miss
            }
        }
    }

    /// Remove a value from the cache
    pub fn remove(&self, key: &K) -> Option<V> {
        let mut store = self.store.lock().unwrap();
        store.remove(key).map(|entry| entry.value)
    }

    /// Clear all entries from the cache
    pub fn clear(&self) {
        let mut store = self.store.lock().unwrap();
        store.clear();
    }

    /// Get current cache size
    pub fn size(&self) -> usize {
        let store = self.store.lock().unwrap();
        store.len()
    }

    /// Get cache statistics
    pub fn stats(&self) -> CacheStats {
        let stats = self.stats.lock().unwrap();
        stats.clone()
    }

    /// Background task to clean up expired entries
    pub async fn cleanup_expired(&self) {
        loop {
            sleep(Duration::from_secs(60)).await;
            
            let mut store = self.store.lock().unwrap();
            let mut stats = self.stats.lock().unwrap();
            
            let expired_keys: Vec<K> = store
                .iter()
                .filter(|(_, entry)| entry.is_expired())
                .map(|(key, _)| key.clone())
                .collect();

            for key in expired_keys {
                store.remove(&key);
                stats.expirations += 1;
            }
        }
    }
}

impl CacheStats {
    pub fn hit_rate(&self) -> f64 {
        let total = self.hits + self.misses;
        if total == 0 {
            0.0
        } else {
            self.hits as f64 / total as f64
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_cache_insert_and_get() {
        let cache = Cache::new(Duration::from_secs(60), 100);
        cache.insert("key1", "value1");
        
        match cache.get(&"key1") {
            CacheResult::Hit(val) => assert_eq!(val, "value1"),
            _ => panic!("Expected cache hit"),
        }
    }
}

