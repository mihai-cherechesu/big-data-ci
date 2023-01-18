def twoSum(nums, target):
        for i in range(len(nums)):
            for j in range(i + 1, len(nums)):
                if nums[j] == target - nums[i]:
                    return [i, j]

def main():
    result = twoSum([2,7,11,15], 9)
    print(str(result[0]) + " " + str(result[1]))

if __name__ == "__main__":
    main()