#include <algorithm>
#include <iostream>
#include <memory>
#include <stdexcept>
#include <string>
#include <vector>

namespace datastructures {

/**
 * Template class for a dynamic array with automatic resizing
 */
template <typename T> class DynamicArray {
private:
  std::unique_ptr<T[]> data;
  size_t capacity;
  size_t length;

  static constexpr size_t INITIAL_CAPACITY = 10;
  static constexpr double GROWTH_FACTOR = 1.5;

  /**
   * Resize the internal array to the new capacity
   */
  void resize(size_t new_capacity) {
    auto new_data = std::make_unique<T[]>(new_capacity);

    for (size_t i = 0; i < length; ++i) {
      new_data[i] = std::move(data[i]);
    }

    data = std::move(new_data);
    capacity = new_capacity;
  }

  /**
   * Ensure capacity is sufficient for additional elements
   */
  void ensure_capacity(size_t required_capacity) {
    if (required_capacity > capacity) {
      size_t new_capacity = std::max(
          required_capacity, static_cast<size_t>(capacity * GROWTH_FACTOR));
      resize(new_capacity);
    }
  }

public:
  /**
   * Constructor with optional initial capacity
   */
  explicit DynamicArray(size_t initial_capacity = INITIAL_CAPACITY)
      : data(std::make_unique<T[]>(initial_capacity)),
        capacity(initial_capacity), length(0) {}

  /**
   * Copy constructor
   */
  DynamicArray(const DynamicArray &other)
      : data(std::make_unique<T[]>(other.capacity)), capacity(other.capacity),
        length(other.length) {
    for (size_t i = 0; i < length; ++i) {
      data[i] = other.data[i];
    }
  }

  /**
   * Move constructor
   */
  DynamicArray(DynamicArray &&other) noexcept
      : data(std::move(other.data)), capacity(other.capacity),
        length(other.length) {
    other.capacity = 0;
    other.length = 0;
  }

  /**
   * Copy assignment operator
   */
  DynamicArray &operator=(const DynamicArray &other) {
    if (this != &other) {
      data = std::make_unique<T[]>(other.capacity);
      capacity = other.capacity;
      length = other.length;

      for (size_t i = 0; i < length; ++i) {
        data[i] = other.data[i];
      }
    }
    return *this;
  }

  /**
   * Move assignment operator
   */
  DynamicArray &operator=(DynamicArray &&other) noexcept {
    if (this != &other) {
      data = std::move(other.data);
      capacity = other.capacity;
      length = other.length;

      other.capacity = 0;
      other.length = 0;
    }
    return *this;
  }

  /**
   * Add an element to the end of the array
   */
  void push_back(const T &value) {
    ensure_capacity(length + 1);
    data[length++] = value;
  }

  /**
   * Add an element using move semantics
   */
  void push_back(T &&value) {
    ensure_capacity(length + 1);
    data[length++] = std::move(value);
  }

  /**
   * Remove and return the last element
   */
  T pop_back() {
    if (length == 0) {
      throw std::out_of_range("Cannot pop from empty array");
    }
    return std::move(data[--length]);
  }

  /**
   * Access element at index with bounds checking
   */
  T &at(size_t index) {
    if (index >= length) {
      throw std::out_of_range("Index out of bounds");
    }
    return data[index];
  }

  /**
   * Const version of at()
   */
  const T &at(size_t index) const {
    if (index >= length) {
      throw std::out_of_range("Index out of bounds");
    }
    return data[index];
  }

  /**
   * Array subscript operator
   */
  T &operator[](size_t index) { return data[index]; }

  /**
   * Const array subscript operator
   */
  const T &operator[](size_t index) const { return data[index]; }

  /**
   * Get current size of the array
   */
  size_t size() const { return length; }

  /**
   * Check if array is empty
   */
  bool empty() const { return length == 0; }

  /**
   * Clear all elements
   */
  void clear() { length = 0; }

  /**
   * Reserve capacity without changing size
   */
  void reserve(size_t new_capacity) {
    if (new_capacity > capacity) {
      resize(new_capacity);
    }
  }

  /**
   * Get current capacity
   */
  size_t get_capacity() const { return capacity; }
};

} // namespace datastructures
