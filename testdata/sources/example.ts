import React, { useState, useEffect, useCallback, useMemo } from "react";
import axios, { AxiosError } from "axios";

// Types and interfaces
interface User {
  id: string;
  name: string;
  email: string;
  role: "admin" | "user" | "guest";
  createdAt: string;
}

interface PaginationState {
  page: number;
  limit: number;
  total: number;
}

interface UserListProps {
  apiUrl: string;
  onUserSelect?: (user: User) => void;
  initialLimit?: number;
}

interface UserListState {
  users: User[];
  loading: boolean;
  error: string | null;
  pagination: PaginationState;
  searchQuery: string;
}

/**
 * UserList component for displaying and managing users
 */
const UserList: React.FC<UserListProps> = ({
  apiUrl,
  onUserSelect,
  initialLimit = 10,
}) => {
  const [state, setState] = useState<UserListState>({
    users: [],
    loading: true,
    error: null,
    pagination: {
      page: 1,
      limit: initialLimit,
      total: 0,
    },
    searchQuery: "",
  });

  /**
   * Fetch users from the API with current filters
   */
  const fetchUsers = useCallback(async () => {
    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      const params = {
        page: state.pagination.page,
        limit: state.pagination.limit,
        search: state.searchQuery,
      };

      const response = await axios.get<{
        data: User[];
        pagination: PaginationState;
      }>(`${apiUrl}/users`, { params });

      setState((prev) => ({
        ...prev,
        users: response.data.data,
        pagination: response.data.pagination,
        loading: false,
      }));
    } catch (err) {
      const error = err as AxiosError;
      setState((prev) => ({
        ...prev,
        loading: false,
        error: error.message || "Failed to fetch users",
      }));
    }
  }, [
    apiUrl,
    state.pagination.page,
    state.pagination.limit,
    state.searchQuery,
  ]);

  /**
   * Effect to fetch users when dependencies change
   */
  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  /**
   * Handle search input changes with debouncing
   */
  const handleSearchChange = useCallback((query: string) => {
    setState((prev) => ({
      ...prev,
      searchQuery: query,
      pagination: { ...prev.pagination, page: 1 },
    }));
  }, []);

  /**
   * Handle pagination changes
   */
  const handlePageChange = useCallback((newPage: number) => {
    setState((prev) => ({
      ...prev,
      pagination: { ...prev.pagination, page: newPage },
    }));
  }, []);

  /**
   * Handle user selection
   */
  const handleUserClick = useCallback(
    (user: User) => {
      if (onUserSelect) {
        onUserSelect(user);
      }
    },
    [onUserSelect]
  );

  /**
   * Compute total pages
   */
  const totalPages = useMemo(() => {
    return Math.ceil(state.pagination.total / state.pagination.limit);
  }, [state.pagination.total, state.pagination.limit]);

  /**
   * Render loading state
   */
  if (state.loading && state.users.length === 0) {
    return <div className="loading">Loading users...</div>;
  }

  /**
   * Render error state
   */
  if (state.error) {
    return (
      <div className="error">
        <p>Error: {state.error}</p>
        <button onClick={fetchUsers}>Retry</button>
      </div>
    );
  }

  /**
   * Main render
   */
  return (
    <div className="user-list">
      <div className="search-bar">
        <input
          type="text"
          placeholder="Search users..."
          value={state.searchQuery}
          onChange={(e) => handleSearchChange(e.target.value)}
        />
      </div>

      <div className="users">
        {state.users.map((user) => (
          <div
            key={user.id}
            className="user-card"
            onClick={() => handleUserClick(user)}
          >
            <h3>{user.name}</h3>
            <p>{user.email}</p>
            <span className={`role ${user.role}`}>{user.role}</span>
          </div>
        ))}
      </div>

      <div className="pagination">
        <button
          disabled={state.pagination.page === 1}
          onClick={() => handlePageChange(state.pagination.page - 1)}
        >
          Previous
        </button>
        <span>
          Page {state.pagination.page} of {totalPages}
        </span>
        <button
          disabled={state.pagination.page >= totalPages}
          onClick={() => handlePageChange(state.pagination.page + 1)}
        >
          Next
        </button>
      </div>
    </div>
  );
};

export default UserList;
