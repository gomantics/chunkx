package com.example.dataprocessor;

import java.util.*;
import java.util.stream.Collectors;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.time.Instant;

/**
 * DataProcessor handles batch processing of data items with filtering and transformation
 */
public class DataProcessor<T> {
    private final List<T> data;
    private final ExecutorService executor;
    private final int batchSize;
    private final Map<String, ProcessingStats> stats;

    /**
     * Statistics for data processing operations
     */
    public static class ProcessingStats {
        private int processed;
        private int failed;
        private long startTime;
        private long endTime;

        public ProcessingStats() {
            this.processed = 0;
            this.failed = 0;
            this.startTime = Instant.now().toEpochMilli();
        }

        public void incrementProcessed() {
            this.processed++;
        }

        public void incrementFailed() {
            this.failed++;
        }

        public void complete() {
            this.endTime = Instant.now().toEpochMilli();
        }

        public long getDuration() {
            return endTime - startTime;
        }

        public int getProcessed() {
            return processed;
        }

        public int getFailed() {
            return failed;
        }
    }

    /**
     * Constructor for DataProcessor
     */
    public DataProcessor(int batchSize, int threadPoolSize) {
        this.data = new ArrayList<>();
        this.executor = Executors.newFixedThreadPool(threadPoolSize);
        this.batchSize = batchSize;
        this.stats = new HashMap<>();
    }

    /**
     * Add a single item to the processor
     */
    public void addItem(T item) {
        if (item != null) {
            data.add(item);
        }
    }

    /**
     * Add multiple items to the processor
     */
    public void addItems(Collection<T> items) {
        if (items != null) {
            data.addAll(items);
        }
    }

    /**
     * Process data in batches using a transformation function
     */
    public <R> List<R> processBatches(
            String operationId,
            DataTransformer<T, R> transformer) {
        
        ProcessingStats operationStats = new ProcessingStats();
        stats.put(operationId, operationStats);

        List<List<T>> batches = createBatches();
        List<CompletableFuture<List<R>>> futures = new ArrayList<>();

        for (List<T> batch : batches) {
            CompletableFuture<List<R>> future = CompletableFuture.supplyAsync(() -> {
                return processBatch(batch, transformer, operationStats);
            }, executor);
            futures.add(future);
        }

        List<R> results = futures.stream()
                .map(CompletableFuture::join)
                .flatMap(List::stream)
                .collect(Collectors.toList());

        operationStats.complete();
        return results;
    }

    /**
     * Process a single batch of items
     */
    private <R> List<R> processBatch(
            List<T> batch,
            DataTransformer<T, R> transformer,
            ProcessingStats stats) {
        
        List<R> results = new ArrayList<>();
        
        for (T item : batch) {
            try {
                R result = transformer.transform(item);
                if (result != null) {
                    results.add(result);
                    stats.incrementProcessed();
                }
            } catch (Exception e) {
                stats.incrementFailed();
                System.err.println("Error processing item: " + e.getMessage());
            }
        }
        
        return results;
    }

    /**
     * Split data into batches
     */
    private List<List<T>> createBatches() {
        List<List<T>> batches = new ArrayList<>();
        
        for (int i = 0; i < data.size(); i += batchSize) {
            int end = Math.min(i + batchSize, data.size());
            batches.add(new ArrayList<>(data.subList(i, end)));
        }
        
        return batches;
    }

    /**
     * Get processing statistics for an operation
     */
    public ProcessingStats getStats(String operationId) {
        return stats.get(operationId);
    }

    /**
     * Clear all data and reset statistics
     */
    public void clear() {
        data.clear();
        stats.clear();
    }

    /**
     * Shutdown the executor service
     */
    public void shutdown() {
        executor.shutdown();
    }

    /**
     * Get total number of items in the processor
     */
    public int size() {
        return data.size();
    }

    /**
     * Functional interface for data transformation
     */
    @FunctionalInterface
    public interface DataTransformer<T, R> {
        R transform(T item) throws Exception;
    }
}

