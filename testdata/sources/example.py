import asyncio
import aiohttp
from typing import List, Dict, Any, Optional
from dataclasses import dataclass
from datetime import datetime


@dataclass
class APIResponse:
    """Represents an API response with metadata"""
    url: str
    status: int
    data: Optional[Dict[str, Any]] = None
    error: Optional[str] = None
    timestamp: float = 0.0


class AsyncAPIClient:
    """Asynchronous API client with rate limiting and retry logic"""
    
    def __init__(self, base_url: str, max_concurrent: int = 10, timeout: int = 30):
        self.base_url = base_url
        self.max_concurrent = max_concurrent
        self.timeout = aiohttp.ClientTimeout(total=timeout)
        self.session: Optional[aiohttp.ClientSession] = None
        self.results: List[APIResponse] = []
    
    async def __aenter__(self):
        """Context manager entry"""
        self.session = aiohttp.ClientSession(timeout=self.timeout)
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit"""
        if self.session:
            await self.session.close()
    
    async def fetch_endpoint(self, endpoint: str, retries: int = 3) -> APIResponse:
        """Fetch data from a single endpoint with retry logic"""
        url = f"{self.base_url}/{endpoint.lstrip('/')}"
        
        for attempt in range(retries):
            try:
                async with self.session.get(url) as response:
                    data = await response.json()
                    return APIResponse(
                        url=url,
                        status=response.status,
                        data=data,
                        timestamp=datetime.now().timestamp()
                    )
            except aiohttp.ClientError as e:
                if attempt == retries - 1:
                    return APIResponse(
                        url=url,
                        status=0,
                        error=str(e),
                        timestamp=datetime.now().timestamp()
                    )
                await asyncio.sleep(2 ** attempt)  # Exponential backoff
        
        return APIResponse(url=url, status=0, error="Max retries exceeded")
    
    async def fetch_multiple(self, endpoints: List[str]) -> List[APIResponse]:
        """Fetch data from multiple endpoints concurrently"""
        semaphore = asyncio.Semaphore(self.max_concurrent)
        
        async def fetch_with_semaphore(endpoint: str) -> APIResponse:
            async with semaphore:
                return await self.fetch_endpoint(endpoint)
        
        tasks = [fetch_with_semaphore(endpoint) for endpoint in endpoints]
        self.results = await asyncio.gather(*tasks)
        return self.results
    
    def get_successful_results(self) -> List[APIResponse]:
        """Filter and return only successful responses"""
        return [r for r in self.results if r.status == 200 and r.data is not None]
    
    def get_failed_results(self) -> List[APIResponse]:
        """Filter and return only failed responses"""
        return [r for r in self.results if r.error is not None or r.status != 200]


async def main():
    """Example usage of AsyncAPIClient"""
    endpoints = [
        "/users",
        "/posts",
        "/comments",
        "/albums",
        "/photos"
    ]
    
    async with AsyncAPIClient("https://jsonplaceholder.typicode.com", max_concurrent=3) as client:
        results = await client.fetch_multiple(endpoints)
        
        successful = client.get_successful_results()
        failed = client.get_failed_results()
        
        print(f"Successful requests: {len(successful)}")
        print(f"Failed requests: {len(failed)}")
        
        for result in failed:
            print(f"Error fetching {result.url}: {result.error}")


if __name__ == "__main__":
    asyncio.run(main())

