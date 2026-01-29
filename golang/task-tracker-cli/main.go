package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type StatusTask string

const (
	StatusTaskTodo       StatusTask = "todo"
	StatusTaskInProgress StatusTask = "in-progress"
	StatusTaskDone       StatusTask = "done"
)

type Task struct {
	ID          int        `json:"id"`
	Description string     `json:"description"`
	Status      StatusTask `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

const fileName = "tasks.json"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Error: Task description required")
			return
		}
		addTask(os.Args[2])

	case "list":
		status := ""
		if len(os.Args) > 2 {
			status = os.Args[2]
		}
		listTasks(status)

	case "update":
		if len(os.Args) < 4 {
			fmt.Println("Error: Task ID and new description required")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Error: Invalid ID")
			return
		}
		updateTask(id, os.Args[3])

	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Error: Task ID required")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Error: Invalid ID")
			return
		}
		deleteTask(id)

	case "mark-in-progress":
		if len(os.Args) < 3 {
			fmt.Println("Error: Task ID required")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Error: Invalid ID")
			return
		}
		updateStatus(id, "in-progress")

	case "mark-done":
		if len(os.Args) < 3 {
			fmt.Println("Error: Task ID required")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Error: Invalid ID")
			return
		}
		updateStatus(id, "done")

	default:
		printUsage()
	}
}

func loadTasks() ([]Task, error) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}, nil
		}
		return nil, err
	}

	var tasks []Task
	err = json.Unmarshal(file, &tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func saveTasks(tasks []Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fileName, data, 0644)
}

func addTask(description string) {
	tasks, err := loadTasks()
	if err != nil {
		fmt.Println("Error loading tasks:", err)
		return
	}

	maxID := 0
	for _, task := range tasks {
		if task.ID > maxID {
			maxID = task.ID
		}
	}

	newTask := Task{
		ID:          maxID + 1,
		Description: description,
		Status:      StatusTaskTodo,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tasks = append(tasks, newTask)

	if err := saveTasks(tasks); err != nil {
		fmt.Println("Error saving task:", err)
		return
	}
	fmt.Printf("Task added successfully (ID: %d)\n", newTask.ID)
}

func listTasks(filterStatus string) {
	tasks, err := loadTasks()
	if err != nil {
		fmt.Println("Error loading tasks:", err)
		return
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}

	fmt.Printf("%-5s %-20s %-12s %s\n", "ID", "Status", "Created", "Description")
	fmt.Println("---------------------------------------------------------------")

	for _, task := range tasks {
		if filterStatus != "" && task.Status != StatusTask(filterStatus) {
			continue
		}

		dateStr := task.CreatedAt.Format("2006-01-02")
		fmt.Printf("%-5d %-20s %-12s %s\n", task.ID, task.Status, dateStr, task.Description)
	}
}

func updateTask(id int, description string) {
	tasks, err := loadTasks()
	if err != nil {
		fmt.Println("Error loading tasks:", err)
		return
	}

	found := false
	for i, task := range tasks {
		if task.ID == id {
			tasks[i].Description = description
			tasks[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Task with ID %d not found\n", id)
		return
	}

	if err := saveTasks(tasks); err != nil {
		fmt.Println("Error saving tasks:", err)
		return
	}
	fmt.Println("Task updated successfully")
}

func updateStatus(id int, status string) {
	tasks, err := loadTasks()
	if err != nil {
		fmt.Println("Error loading tasks:", err)
		return
	}

	found := false
	for i, task := range tasks {
		if task.ID == id {
			tasks[i].Status = StatusTask(status)
			tasks[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Task with ID %d not found\n", id)
		return
	}

	if err := saveTasks(tasks); err != nil {
		fmt.Println("Error saving tasks:", err)
		return
	}
	fmt.Println("Task status updated successfully")
}

func deleteTask(id int) {
	tasks, err := loadTasks()
	if err != nil {
		fmt.Println("Error loading tasks:", err)
		return
	}

	var newTasks []Task
	found := false
	for _, task := range tasks {
		if task.ID == id {
			found = true
			continue
		}
		newTasks = append(newTasks, task)
	}

	if !found {
		fmt.Printf("Task with ID %d not found\n", id)
		return
	}

	if err := saveTasks(newTasks); err != nil {
		fmt.Println("Error saving tasks:", err)
		return
	}
	fmt.Println("Task deleted successfully")
}

func printUsage() {
	fmt.Println("Usage: task-cli [command] [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  add \"description\"            	Add a new task")
	fmt.Println("  list [status]                	List tasks (optional: done, todo, in-progress)")
	fmt.Println("  update [id] \"description\"    	Update a task description")
	fmt.Println("  delete [id]                  	Delete a task")
	fmt.Println("  mark-in-progress [id]        	Mark task as in-progress")
	fmt.Println("  mark-done [id]               	Mark task as done")
}
