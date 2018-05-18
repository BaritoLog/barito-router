/* Collection of helper for slice of string */
package strslice

func Contain(slice []string, s string)  bool{
  for _, item := range slice {
      if item == s {
        return true
      }
  }
  
  return false
}
