#include <stdlib.h>
#include <stdio.h>

int *twoSum(int *nums, int numsSize, int target){
    int *returnValues = malloc(2 * sizeof(int));

    for (int i = 0; i < numsSize - 1; i++) {
        for (int j = i + 1; j < numsSize; j++) {
            if (nums[i] + nums[j] == target) {
                returnValues[0] = i;
                returnValues[1] = j;
                break;
            }
        }
    }
    
    return returnValues;
}

int main() {
    int nums[] = {2, 7, 11, 15};
    int nums_size = 4;
    int target = 9;
    int *result;
    result = twoSum(nums, nums_size, target);
    printf("%d %d\n", result[0], result[1]);
    return 0;
}