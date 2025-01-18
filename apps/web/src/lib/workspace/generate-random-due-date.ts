function generateRandomDueDate() {
  const today = new Date();
  const futureDate = new Date();
  futureDate.setDate(today.getDate() + Math.floor(Math.random() * 30));
  return futureDate.toLocaleDateString();
}

export default generateRandomDueDate;
